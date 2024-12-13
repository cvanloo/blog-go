package gen

import (
	"strings"
	"io"
	"net/http"
	"time"
	"fmt"
	"bytes"
	"html/template"
	"embed"
	"log"
)

//go:embed html
var htmls embed.FS

var pages = Template{template.New("")}

func init() {
	pages.Funcs(template.FuncMap{
		"Render": Render,
	})

	template.Must(pages.ParseFS(htmls, "html/*.gohtml"))
	log.Println(pages.DefinedTemplates())
}

func Render(element Renderable) (template.HTML, error) {
	return element.Render()
}

type (
	Template struct {
		*template.Template
	}
	Renderable interface {
		Render() (template.HTML, error)
	}
)

func (t *Template) Execute(w io.Writer, name string, data any) error {
	return t.Template.ExecuteTemplate(w, name, data)
}

func String(blog *Blog) (string, error) {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "entry.gohtml", blog)
	return bs.String(), err
}

func Handler(blog *Blog, onError func(error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := pages.Execute(w, "entry.gohtml", blog)
		if err != nil {
			onError(err)
		}
	}
}

type (
	Blog struct {
		Author Author
		Lang string
		Title, AltTitle string
		Published Revision
		EstReading int
		Tags []Tag
		Series *Series
		EnableRevisionWarning bool
		TOC TableOfContents
		Abstract []Renderable
		Sections []Section
		Relevant *RelevantBox
	}
	Author struct {
		Name, Email string
	}
	Revision struct {
		Published time.Time
		Revised *time.Time
	}
	Tag string
	Series struct {
		Prev, Next *SeriesItem
	}
	SeriesItem struct {
		Title, Link string
	}
	TableOfContents struct {
		Sections []TOCSection
	}
	TOCSection struct {
		ID, Heading string
		NextLevel []TOCSection
	}
	Section struct {
		Level int
		Heading string
		Content []Renderable
	}
	Paragraph struct {
		Content []Renderable
	}
	Text string
	Bold string
	Italic string
	Mono string
	Link struct {
		Href, Text string
	}
	CodeBlock struct {
		Lines []string
	}
	Sidenote struct {
		ID string
		Word, Content string
	}
	Picture struct {
		Name, Title, Alt string
	}
	Blockquote struct {
		QuoteText string
		Author, Source string
	}
	RelevantBox struct {
		Heading string
		Articles []ReadingItem
	}
	ReadingItem struct {
		Link, Title, AuthorLink, Author, Abstract string
		Date time.Time
	}
)

func (b *Blog) Canonical() string {
	return "https://blog.vanloo.ch/blog" // @todo
}

func (b *Blog) RelMe() string {
	return "https://tech.lgbt/@attaboy" // @todo
}

func (b *Blog) FediCreator() string {
	return "@attaboy@tech.lgbt" // @todo
}

func (b *Blog) FullTitle() string {
	return fmt.Sprintf("%s&mdash;%s", b.Title, b.AltTitle)
}

func (b *Blog) FirstSectionID() string {
	return b.Sections[0].ID()
}

func (b *Blog) FirstSectionName() string {
	return b.Sections[0].Heading
}

func (r Revision) HasRevision() bool {
	return r.Revised != nil
}

func (r Revision) String() string {
	const timeFormat = "Mon, 2 Jan 2006"
	if r.Revised != nil {
		return fmt.Sprintf("%s (revised %s)", r.Published.Format(timeFormat), r.Revised.Format(timeFormat))
	}
	return fmt.Sprintf("%s", r.Published.Format(timeFormat))
}

func (t Tag) String() string {
	return string(t)
}

func (b *Blog) IsPartOfSeries() bool {
	return b.Series != nil
}

func (s *Series) HasPrev() bool {
	return s.Prev != nil
}

func (s *Series) HasNext() bool {
	return s.Next != nil
}

func (b *Blog) LastRevision() time.Time {
	if b.Published.Revised != nil {
		return *b.Published.Revised
	}
	return b.Published.Published
}

func (b *Blog) ShowLongTimeSinceRevisedWarning() bool {
	const threeYears = 3 * 365 * 24 * time.Hour
	return b.EnableRevisionWarning && time.Since(b.LastRevision()) > threeYears
}

func (t *TableOfContents) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "toc.gohtml", t)
	return template.HTML(bs.String()), err
}

func (s *TOCSection) HasNextLevel() bool {
	return len(s.NextLevel) > 0
}

func (s Section) ID() string {
	return strings.ReplaceAll(strings.ToLower(s.Heading), " ", "-")
}

func (s Section) SectionLevel1() bool {
	return s.Level == 1
}

func (s Section) SectionLevel2() bool {
	return s.Level == 2
}

func (s Section) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "section.gohtml", s)
	return template.HTML(bs.String()), err
}

func (p Paragraph) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "paragraph.gohtml", p)
	return template.HTML(bs.String()), err
}

func (t Text) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "text.gohtml", t)
	return template.HTML(bs.String()), err
}

func (l Link) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "link.gohtml", l)
	return template.HTML(bs.String()), err
}

func (l Link) Target() string {
	return "_blank" // @todo
}

func (cb CodeBlock) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "code-block.gohtml", cb)
	return template.HTML(bs.String()), err
}

func (sn Sidenote) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "sidenote.gohtml", sn)
	return template.HTML(bs.String()), err
}

func (p Picture) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "picture.gohtml", p)
	return template.HTML(bs.String()), err
}

func (b Blockquote) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "blockquote.gohtml", b)
	return template.HTML(bs.String()), err
}

func (b *Blog) ShowRelevantSection() bool {
	return b.Relevant != nil
}

func (r ReadingItem) FormatDate() string {
	return r.Date.Format("2006-01-02")
}

func (b *Blog) CopyrightYear() string {
	return time.Now().Format("2006")
}

func (b *Blog) ObfuscatedEmail() string {
	return b.Author.Email // @todo
}

func (b *Blog) PublishedFull() string {
	return b.Published.Published.Format(time.RFC3339)
}

func (b *Blog) RevisedFull() string {
	return b.Published.Revised.Format(time.RFC3339)
}

func (b *Blog) AbstractShort() string {
	return "" // @todo
}

func (b *Blog) HasAbstract() bool {
	return len(b.Abstract) > 0
}

func (a Author) String() string {
	return a.Name
}

func (b Bold) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "bold.gohtml", b)
	return template.HTML(bs.String()), err
}

func (i Italic) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "italic.gohtml", i)
	return template.HTML(bs.String()), err
}

func (m Mono) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "mono.gohtml", m)
	return template.HTML(bs.String()), err
}
