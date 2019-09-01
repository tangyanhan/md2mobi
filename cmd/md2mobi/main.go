package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"gopkg.in/yaml.v2"

	"github.com/tangyanhan/md2mobi/mobi"
)

// Config for convert
type Config struct {
	Title  string `yaml:"title,omitempty"`
	Author string `yaml:"author,omitempty"`
	Cover  string `yaml:"cover"`
	Thumb  string `yaml:"thumb"`
	// Max level for a chapter. 0: entire file will be added as a single chapter. 1: <h1> elements will be used as maximum chapter
	MaxLevel int `yaml:"max_level"`
	// Filename mapping to chapter name, if not specified, a.md will use a as chapter name
	Names map[string]string `yaml:"names"`
}

func main() {
	var (
		dirPath    string
		filePath   string
		outputPath string
		configFile string
	)
	flag.StringVar(&dirPath, "d", "", "Convert all *.md files under dir into a single mobi")
	flag.StringVar(&filePath, "f", "", "Convert a single markdown file into mobi")
	flag.StringVar(&outputPath, "o", "out.mobi", "The mobi file path to be generated")
	flag.StringVar(&configFile, "c", "", "Config file for detailed generation")
	flag.Parse()
	if dirPath == "" && filePath == "" {
		flag.Usage()
		os.Exit(1)
	}
	cfg := Config{
		Title:  "DefaultTitle",
		Author: "Ethan Tang",
	}
	if configFile != "" {
		raw, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Fatalf("Failed to read config from %s: %v", configFile, err)
		}
		if err = yaml.Unmarshal(raw, &cfg); err != nil {
			log.Fatalf("Failed to unmarshal config file from %s:%v", configFile, err)
		}
	}
	var err error
	if dirPath != "" {
		err = convertDirToMobi(dirPath, outputPath)
	} else {
		w, err := mobi.NewWriter(outputPath)
		if err != nil {
			panic(err)
		}
		fileName := path.Base(filePath)
		ext := path.Ext(fileName)
		title := strings.TrimSuffix(fileName, ext)
		if cfg.Title == "" {
			cfg.Title = title
		}
		w.Title(cfg.Title)
		w.Compression(mobi.CompressionNone) // LZ77 compression is also possible using  mobi.CompressionPalmDoc
		// Add cover image
		if cfg.Cover != "" && cfg.Thumb != "" {
			w.AddCover(cfg.Cover, cfg.Thumb)
		}

		// Meta data
		w.NewExthRecord(mobi.EXTH_DOCTYPE, "EBOK")
		w.NewExthRecord(mobi.EXTH_AUTHOR, cfg.Author)

		err = fileToMobi(w, filePath, title, cfg)
		w.Write()
	}
	if err != nil {
		log.Fatalln("Failed to convert mobi:", err.Error())
	}
}

func convertDirToMobi(dirPath, mobiPath string) error {
	return nil
}

func fileToMobi(w *mobi.MobiWriter, filePath, fileName string, cfg Config) error {
	chapName, ok := cfg.Names[fileName]
	if !ok {
		chapName = fileName
	}
	rawMD, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read markdown from %s:%v", filePath, err)
	}
	docRoot := markdown.Parse(rawMD, nil)
	opts := html.RendererOptions{
		Flags: html.CommonFlags,
	}
	renderer := html.NewRenderer(opts)
	chap := w.NewChapter(chapName, nil)
	renderToMobi(chap, docRoot, renderer, cfg.MaxLevel)
	return nil
}

func renderToMobi(chap mobi.Chapter, node ast.Node, renderer markdown.Renderer, maxLevel int) {
	root, ok := node.(*ast.Document)
	if ok {
		chap.SetHTML(root.Content)
	} else {
		nodeData, ok := node.(*ast.Heading)

		if !ok || nodeData.Level > maxLevel {
			raw := markdown.Render(node, renderer)
			log.Printf("Final: node:%v %s", reflect.TypeOf(node), string(raw))
			sub := mobi.NewChapter("", raw)
			chap.AddSubChapter(sub)
			return
		}

		sub := mobi.NewChapter("", nodeData.Content)
		chap = chap.AddSubChapter(sub)
	}

	var title, content string
	// TODO: Logic here is messed up, have to think twice, or more
	for _, child := range node.GetChildren() {
		txt, ok := node.(*ast.Text)
		if ok {
			title = string(txt.Content)
			chap.SetTitle(string(txt.Content))
			continue
		}
		para, ok := node.(*ast.Paragraph)
		if ok {
			content = string(para.Content)
			chap.SetHTML(para.Content)
			continue
		}
		raw := markdown.Render(node, renderer)
		fmt.Printf("Node Type=%v NodeContent=%s", reflect.TypeOf(child), string(raw))
		renderToMobi(chap, child, renderer, maxLevel)
	}
	log.Printf("Chap=%s Content=%s", title, content)
}
