package gen

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"net/url"

	. "github.com/cvanloo/blog-go/assert"
)

//go:embed html
var htmls embed.FS

var (
	pages = Template{Template: template.New("")}
	site = Site{
		Address: Must(url.Parse("https://blog.vanloo.ch")), // @todo
	}
)

func init() {
	pages.Funcs(template.FuncMap{
		"Render": Render,
		"GetBlog": func() *Blog {
			return pages.blog
		},
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
		blog *Blog
	}
	Renderable interface {
		Render() (template.HTML, error)
	}
	StringRenderable interface {
		Renderable
		Text() string
	}
)

func (t *Template) Execute(w io.Writer, name string, data any) error {
	return t.Template.ExecuteTemplate(w, name, data)
}

func WriteBlog(w io.Writer, blog *Blog) error {
	pages.blog = blog
	return pages.Execute(w, "entry.gohtml", blog)
}

func String(blog *Blog) (string, error) {
	bs := &bytes.Buffer{}
	pages.blog = blog
	err := pages.Execute(bs, "entry.gohtml", blog)
	return bs.String(), err
}

func Handler(blog *Blog, onError func(error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pages.blog = blog
		err := pages.Execute(w, "entry.gohtml", blog)
		if err != nil {
			onError(err)
		}
	}
}

type (
	Attributeable interface {
		SetID(id string)
		GetID() string
		SetAttr(key, val string)
		GetAttr(key string) (string, bool)
	}
	Attributes struct {
		ID string
		Fields map[string]string
	}
	Blog struct { // @todo: make Meta struct{}, // @todo: mandatory keys?
		UrlPath string
		Author Author
		Lang string
		Title, AltTitle StringRenderable
		Description string
		Published Revision
		EstReading int
		Tags []Tag
		Series *Series
		EnableRevisionWarning bool
		TOC TableOfContents
		Abstract StringRenderable
		Sections []Section
		Relevant *RelevantBox
	}
	Site struct {
		Address *url.URL // e.g. https://blog.vanloo.ch
	}
	Author struct {
		Name StringRenderable
		Email StringRenderable
		RelMe StringRenderable // https://tech.lgbt/@attaboy
		FediCreator StringRenderable // @attaboy@tech.lgbt
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
		Title StringRenderable
		Link string
	}
	TableOfContents struct {
		Sections []TOCSection
	}
	TOCSection struct {
		ID string
		Heading StringRenderable
		NextLevel []TOCSection
	}
	Section struct {
		Attributes
		Level int
		Heading StringRenderable
		Content []Renderable
	}
	Paragraph struct {
		Content StringRenderable
	}
	Text string
	Mono string
	EscapedString string
	StringOnlyContent []StringRenderable
	Strong struct {
		StringOnlyContent
	}
	Emphasis struct {
		StringOnlyContent
	}
	EmphasisStrong struct {
		StringOnlyContent
	}
	EnquoteDouble struct {
		StringOnlyContent
	}
	EnquoteAngled struct {
		StringOnlyContent
	}
	Strikethrough struct {
		StringOnlyContent
	}
	Marker struct {
		StringOnlyContent
	}
	Link struct {
		Ref, Href string
		Name StringRenderable
	}
	CodeBlock struct {
		Attributes
		Lines []string
	}
	Sidenote struct {
		Ref string
		// @todo: For the title attribute we can't have <b> and stuff...
		Word, Content StringRenderable
	}
	Image struct {
		Name string
		Title, Alt StringRenderable
	}
	Blockquote struct {
		QuoteText, Author, Source StringRenderable
	}
	HorizontalRule struct{}
	RelevantBox struct {
		Heading StringRenderable
		Articles []ReadingItem
	}
	ReadingItem struct {
		Link, AuthorLink string
		Title, Author, Abstract StringRenderable
		Date time.Time
	}
)

func (a Attributes) SetAttr(key, val string) {
	if a.Fields == nil {
		a.Fields = map[string]string{}
	}
	a.Fields[key] = val
}

func (a Attributes) GetAttr(key string) (string, bool) {
	val, ok := a.Fields[key]
	return val, ok
}

func (a Attributes) SetID(id string) {
	a.ID = id
}

func (a Attributes) GetID() string {
	return a.ID
}

func (soc StringOnlyContent) Render() (template.HTML, error) {
	return template.HTML(soc.Text()), nil
}

func (soc StringOnlyContent) Text() string {
	var builder strings.Builder
	for _, s := range soc {
		builder.WriteString(s.Text())
	}
	return builder.String()
}

func (soc StringOnlyContent) String() string {
	log.Printf("warning: String() called on StringOnlyContent, you probably want to use Render from within the html template: %#v", soc)
	return fmt.Sprintf("%#v", soc)
}

func (t Text) Render() (template.HTML, error) {
	return template.HTML(t), nil
}

func (t Text) Text() string {
	return string(t)
}

func (s Strong) Render() (template.HTML, error) {
	return template.HTML(s.Text()), nil
}

func (s Strong) Text() string {
	return fmt.Sprintf("<strong>%s</strong>", s.StringOnlyContent.Text())
}

func (e Emphasis) Render() (template.HTML, error) {
	return template.HTML(e.Text()), nil
}

func (e Emphasis) Text() string {
	return fmt.Sprintf("<em>%s</em>", e.StringOnlyContent.Text())
}

func (e EmphasisStrong) Render() (template.HTML, error) {
	return template.HTML(e.Text()), nil
}

func (e EmphasisStrong) Text() string {
	return fmt.Sprintf("<em><strong>%s</strong></em>", e.StringOnlyContent.Text())
}

func (m Mono) Render() (template.HTML, error) {
	return template.HTML(m.Text()), nil
}

func (m Mono) Text() string {
	return fmt.Sprintf("<code>%s</code>", m)
}

func (q EnquoteDouble) Render() (template.HTML, error) {
	return template.HTML(q.Text()), nil
}

func (q EnquoteDouble) Text() string {
	return fmt.Sprintf("&ldquo;%s&rdquo;", q.StringOnlyContent.Text())
}

func (q EnquoteAngled) Render() (template.HTML, error) {
	return template.HTML(q.Text()), nil
}

func (q EnquoteAngled) Text() string {
	return fmt.Sprintf("&laquo;%s&raquo;", q.StringOnlyContent.Text())
}

func (s Strikethrough) Text() string {
	return fmt.Sprintf("<s>%s</s>", s.StringOnlyContent.Text())
}

func (s Strikethrough) Render() (template.HTML, error) {
	return template.HTML(s.Text()), nil
}

func (m Marker) Text() string {
	return fmt.Sprintf("<mark>%s</mark>", m.StringOnlyContent.Text())
}

func (m Marker) Render() (template.HTML, error) {
	return template.HTML(m.Text()), nil
}

func (l Link) Render() (template.HTML, error) {
	return template.HTML(l.Text()), nil
}

func (l Link) Target() string {
	// @todo: check if it's a link referring to a section in the same blog post.
	//    then add a css class, so that we can show an arrow-up or arrow-down
	//    (depending on the relative position of the link and the section it points to)
	href := Must(url.Parse(l.Href))
	if href.Host == site.Address.Host {
		return "_self"
	}
	return "_blank"
}

func (l Link) Text() string {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "link.gohtml", l)
	_ = err // @todo
	return strings.TrimSpace(bs.String())
}

func (l Link) NameOrHref() string {
	if l.Name != nil {
		return l.Name.Text()
	}
	return l.Href
}

func (e EscapedString) Render() (template.HTML, error) {
	return template.HTML(e.Text()), nil
}

func (e EscapedString) Text() string {
	return string(e)
}

const (
	AmpNoBreakSpace EscapedString = "&nbsp;"
	AmpEmDash EscapedString = "&mdash;"
	AmpEnDash EscapedString = "&ndash;"
	AmpHyphen EscapedString = "&hyphen;"
	AmpLeftDoubleQuote EscapedString = "&ldquo;"
	AmpRightDoubleQuote EscapedString = "&rdquo;"
	AmpLeftAngledQuote EscapedString = "&laquo;"
	AmpRightAngledQuote EscapedString = "&raquo;"
	AmpEllipsis EscapedString = "…"
	AmpPrime EscapedString = "&prime;"
	AmpDoublePrime EscapedString = "&Prime;"
	AmpTripplePrime EscapedString = "&tprime;"
	AmpQuadruplePrime EscapedString = "&qprime;"
	AmpReversedPrime EscapedString = "&bprime;"
)

func (sn Sidenote) Render() (template.HTML, error) {
	return template.HTML(sn.Text()), nil
}

func (sn Sidenote) Text() string {
	bs := &bytes.Buffer{}
	PanicIf(pages.Execute(bs, "sidenote.gohtml", sn)) // @todo: how do we do error handling here? I guess Text() must also return an error after all?
	return strings.TrimSpace(bs.String())
}

func (sn Sidenote) ID() string {
	// @todo: implement incrementing counter
	return ""
}

func (b *Blog) Canonical() string {
	path := b.UrlPath
	Assert(len(path) > 0, "must specify a url path")
	//if path == "" {
	//	log.Println("warning: url path not set, deriving from title")
	//	path = strings.ReplaceAll(strings.ToLower(b.Title.Text()), " ", "-")
	//}
	// @todo: or based on the file name?
	Assert(site.Address != nil, "must specify a site address")
	//if site == "" {
	//	log.Println("warning: site address not set, defaulting to http://localhost")
	//	site = "http://localhost"
	//}
	return fmt.Sprintf("%s://%s/%s", site.Address.Scheme, site.Address.Host, path)
}

func (b *Blog) FirstSectionID() string {
	Assert(len(b.Sections) > 0, "blog must consist of at least one section")
	return b.Sections[0].ID()
}

func (b *Blog) FirstSectionName() string {
	Assert(len(b.Sections) > 0, "blog must consist of at least one section")
	return b.Sections[0].Heading.Text()
}

func (b *Blog) LastRevision() time.Time {
	if b.Published.HasRevision() {
		return *b.Published.Revised
	}
	return b.Published.Published
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

func (b *Blog) ShowLongTimeSinceRevisedWarning() bool {
	const threeYears = 3 * 365 * 24 * time.Hour // doesn't have to be exact, or even care about time zones and stuff
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
	// @todo: check and error if IDs collide (here? and/or in parser?)
	if len(s.Attributes.ID) != 0 {
		return s.Attributes.ID
	}
	return strings.ReplaceAll(strings.ToLower(s.Heading.Text()), " ", "-")
}

func (s Section) SectionLevel1() bool {
	return s.Level == 1
}

// @todo: separate template, separate type
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

func (cb CodeBlock) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "code-block.gohtml", cb)
	return template.HTML(bs.String()), err
}

func (i Image) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "image.gohtml", i)
	return template.HTML(bs.String()), err
}

func (b Blockquote) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "blockquote.gohtml", b)
	return template.HTML(bs.String()), err
}

func (hr HorizontalRule) Render() (template.HTML, error) {
	return template.HTML("\n<hr>\n"), nil
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

func (b *Blog) ObfuscatedEmail() template.HTML {
	const (
		janetStart = `janet -e '(print (string/from-bytes (splice (map (fn [c] (if (<= 97 c 122) (+ 97 (mod (+ c -97 13) 26)) c)) &quot;`
		janetEnd = `&quot;))))'`
	)
	rot13 := func(text string) string {
		out := []rune(text)
		for i, r := range out {
			if r >= 'a' && r <= 'z' {
				out[i] = ((r - 'a' + 13) % 26) + 'a'
			}
		}
		return string(out)
	}
	return template.HTML(janetStart + rot13(b.Author.Email.Text()) + janetEnd)
}

func (b *Blog) PublishedFull() string {
	return b.Published.Published.Format(time.RFC3339)
}

func (b *Blog) RevisedFull() string {
	Assert(b.Published.HasRevision(), "must check HasRevision to know if it is safe to access Revised")
	return b.Published.Revised.Format(time.RFC3339)
}

func (b *Blog) AbstractShort() StringRenderable {
	return b.Abstract // @todo: cut short? one - two sentences?
}

func (b *Blog) HasAbstract() bool {
	return b.Abstract != nil
}

var (
	currentID = 0
)

func NextID() int {
	currentID++
	return currentID
}
