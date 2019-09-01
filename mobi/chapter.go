package mobi

import "bytes"

// Chapter for mobi
type Chapter interface {
	AddSubChapter(subChap Chapter) Chapter
	SetTitle(title string)
	SetHTML(text []byte)
	SubChapterCount() int
}

type mobiChapter struct {
	Id           int
	Parent       int
	Title        string
	RecordOffset int
	LabelOffset  int
	Len          int
	Html         []uint8
	SubChapters  []*mobiChapter
}

func (w *MobiWriter) NewChapter(title string, text []byte) Chapter {
	w.chapters = append(w.chapters, mobiChapter{Id: w.chapterCount, Title: title, Html: minimizeHTML(text)})
	w.chapterCount++
	return &w.chapters[len(w.chapters)-1]
}

func NewChapter(title string, text []byte) Chapter {
	return &mobiChapter{Title: title, Html: minimizeHTML(text)}
}

func (w *mobiChapter) AddSubChapter(subChap Chapter) Chapter {
	sub, ok := subChap.(*mobiChapter)
	if !ok {
		panic("cannot cast as mobiChapter")
	}
	sub.Parent = w.Id
	w.SubChapters = append(w.SubChapters, sub)
	return w
}

func (w *mobiChapter) SetTitle(title string) {
	w.Title = title
}

func (w *mobiChapter) SetHTML(text []byte) {
	w.Html = minimizeHTML(text)
}

func (w *mobiChapter) SubChapterCount() int {
	return len(w.SubChapters)
}

func (w *mobiChapter) generateHTML(out *bytes.Buffer) {
	//Add check for unsupported HTML tags, characters, clean up HTML
	w.RecordOffset = out.Len()
	Len0 := out.Len()
	//fmt.Printf("Offset: --- %v %v \n", w.Offset, w.Title)
	out.WriteString("<h1>" + w.Title + "</h1>")
	out.Write(w.Html)
	out.WriteString("<mbp:pagebreak/>")
	w.Len = out.Len() - Len0
	for i, _ := range w.SubChapters {
		w.SubChapters[i].generateHTML(out)
	}
}
