package page

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	. "github.com/cvanloo/blog-go/assert"
	"github.com/cvanloo/blog-go/markup/parser"
)

//go:embed post
var postTemplates embed.FS

var (
	post = Template{Template: template.New("")}
)

func init() {
	post.Funcs(template.FuncMap{
		"Render":         Render,
		"MakeUniqueID":   MakeUniqueID,
		"ObfuscateText":  ObfuscateText,
		"CopyrightYear":  CopyrightYear,
		"CopyrightYears": CopyrightYears,
		"UrlEscapeLower": UrlEscapeLower,
	})
	template.Must(post.ParseFS(postTemplates, "post/*.gohtml"))
	log.Printf("post: %s", post.DefinedTemplates())
}

type (
	Attributes map[string]string
	Post       struct {
		MakePublish           bool
		Site                  Site
		UrlPath               string
		Author                Author
		Lang                  string
		Title, AltTitle       StringRenderable
		Description           string
		Abstract              []Renderable
		TopLevelContent       []Renderable
		Published             Revision
		EstReading            int
		WordCount             int
		Tags                  []Tag
		Series                *Series
		EnableRevisionWarning bool
		TOC                   TableOfContents
		Sections              []Section
		Relevant              *RelevantBox // @todo: implement: need html custom functions first
	}
	Author struct {
		Name        StringRenderable
		Email       StringRenderable
		RelMe       StringRenderable // https://tech.lgbt/@attaboy
		FediCreator StringRenderable // @attaboy@tech.lgbt
	}
	Series struct {
		Name       StringRenderable
		Link       string
		Prev, Next *SeriesItem
	}
	SeriesItem struct {
		Title StringRenderable
		Link  string
	}
	TableOfContents struct {
		Sections []TOCSection
	}
	TOCSection struct {
		ID        string
		Heading   StringRenderable
		NextLevel []TOCSection
	}
	Section struct {
		Attributes
		Level   int
		Heading StringSanitizedRenderable
		Content []Renderable
	}
	Paragraph struct {
		Content StringRenderable
	}
	Text              string
	Mono              string
	EscapedString     string
	StringOnlyContent []StringRenderable
	Strong            struct {
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
		Name StringRenderable
		Href string
	}
	CodeBlock struct {
		Attributes
		Lines []string
	}
	Sidenote struct {
		// @todo: For the title attribute we can't have <b> and stuff...
		Word, Content StringRenderable
	}
	Note struct {
		Type string
		Content []Renderable
	}
	Ruby struct {
		Kanji, Furigana StringRenderable
	}
	Image struct {
		Name       string
		Title, Alt StringRenderable
	}
	Video struct {
		Name       string
		Title, Alt StringRenderable
	}
	Blockquote struct {
		Attributes
		QuoteText, Author, Source StringRenderable
	}
	HorizontalRule struct{}
	LineBreak      struct{}
	RelevantBox    struct {
		Heading  StringRenderable
		Articles []ReadingItem
	}
	ReadingItem struct {
		Link, AuthorLink string
		Title, Author    StringRenderable
		Abstract         []Renderable
		Date             time.Time
	}
)

func WritePost(w io.Writer, p Post) error {
	p.Site = SiteInfo // @todo
	return post.Execute(w, "post.gohtml", p)
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

func (soc StringOnlyContent) SanitizedText() string {
	var builder strings.Builder
	for _, s := range soc {
		if s, ok := s.(StringSanitizedRenderable); ok {
			builder.WriteString(s.SanitizedText())
		}
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

func (t Text) SanitizedText() string {
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
	bs := &bytes.Buffer{}
	PanicIf(post.Execute(bs, "mono.gohtml", m))
	return strings.TrimSpace(bs.String())
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

func (n Note) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	PanicIf(post.Execute(bs, "note.gohtml", n))
	return template.HTML(bs.String()), nil
}

func (n Note) TypeOrDefault() string {
	if n.Type == "" {
		return "info"
	}
	return strings.ToLower(n.Type)
}

func (r Ruby) Text() string {
	bs := &bytes.Buffer{}
	PanicIf(post.Execute(bs, "ruby.gohtml", r))
	return strings.TrimSpace(bs.String())
}

func (r Ruby) Render() (template.HTML, error) {
	return template.HTML(r.Text()), nil
}

func (l LineBreak) Text() string {
	return "<br>"
}

func (l LineBreak) Render() (template.HTML, error) {
	return template.HTML(l.Text()), nil
}

func (l Link) Render() (template.HTML, error) {
	return template.HTML(l.Text()), nil
}

func (l Link) Target() (string, error) {
	// @todo: check if it's a link referring to a section in the same blog post.
	//    then add a css class, so that we can show an arrow-up or arrow-down
	//    (depending on the relative position of the link and the section it points to)
	href, err := url.Parse(l.Href)
	if err != nil {
		return "_blank", err
	}
	if href.Host == SiteInfo.Address.Host {
		return "_self", nil
	}
	return "_blank", nil
}

func (l Link) Text() string {
	bs := &bytes.Buffer{}
	PanicIf(post.Execute(bs, "link.gohtml", l))
	return strings.TrimSpace(bs.String())
}

func (l Link) NameOrHref() template.HTML {
	if l.Name != nil {
		return template.HTML(l.Name.Text())
	}
	return template.HTML(l.Href)
}

func (e EscapedString) Render() (template.HTML, error) {
	return template.HTML(e.Text()), nil
}

func (e EscapedString) Text() string {
	return string(e)
}

const (
	AmpNoBreakSpace                 EscapedString = "&nbsp;"
	AmpEmDash                       EscapedString = "&mdash;"
	AmpEnDash                       EscapedString = "&ndash;"
	AmpHyphen                       EscapedString = "&hyphen;"
	AmpLeftSingleQuote              EscapedString = "&lsquo;"
	AmpRightSingleQuoteOrApostrophe EscapedString = "&rsquo;"
	AmpLeftDoubleQuote              EscapedString = "&ldquo;"
	AmpRightDoubleQuote             EscapedString = "&rdquo;"
	AmpLeftAngledQuote              EscapedString = "&laquo;"
	AmpRightAngledQuote             EscapedString = "&raquo;"
	AmpEllipsis                     EscapedString = "…"
	AmpPrime                        EscapedString = "&prime;"
	AmpDoublePrime                  EscapedString = "&Prime;"
	AmpTripplePrime                 EscapedString = "&tprime;"
	AmpQuadruplePrime               EscapedString = "&qprime;"
	AmpReversedPrime                EscapedString = "&bprime;"
)

func (sn Sidenote) Render() (template.HTML, error) {
	return template.HTML(sn.Text()), nil
}

func (sn Sidenote) Text() string {
	bs := &bytes.Buffer{}
	PanicIf(post.Execute(bs, "sidenote.gohtml", sn))
	return strings.TrimSpace(bs.String())
}

func (p Post) Canonical() string {
	path := p.UrlPath
	return fmt.Sprintf("%s://%s/%s", SiteInfo.Address.Scheme, SiteInfo.Address.Host, path)
}

func (p Post) FirstSectionID() string {
	Assert(len(p.Sections) > 0, "blog must consist of at least one section")
	return p.Sections[0].ID()
}

func (p Post) FirstSectionName() string {
	Assert(len(p.Sections) > 0, "blog must consist of at least one section")
	return p.Sections[0].Heading.Text()
}

func (p Post) LastRevision() time.Time {
	if p.Published.HasRevision() {
		return *p.Published.Revised
	}
	return p.Published.Published
}

func (p Post) IsPartOfSeries() bool {
	return p.Series != nil
}

func (s *Series) HasPrev() bool {
	return s.Prev != nil
}

func (s *Series) HasNext() bool {
	return s.Next != nil
}

func (p Post) ShowLongTimeSinceRevisedWarning() bool {
	const threeYears = 3 * 365 * 24 * time.Hour // doesn't have to be exact, or even care about time zones and stuff
	return p.EnableRevisionWarning && time.Since(p.LastRevision()) > threeYears
}

func (t TableOfContents) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := post.Execute(bs, "toc.gohtml", t)
	return template.HTML(bs.String()), err
}

func (s *TOCSection) HasNextLevel() bool {
	return len(s.NextLevel) > 0
}

func (s Section) ID() string {
	// @todo: .Text() will give us something like `it&rsquo;s` but we actually need `it's`
	if id, ok := s.Attributes["id"]; ok {
		return id
	}
	return strings.ReplaceAll(strings.ToLower(s.Heading.SanitizedText()), " ", "-")
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
	err := post.Execute(bs, "section.gohtml", s)
	return template.HTML(bs.String()), err
}

func (p Paragraph) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := post.Execute(bs, "paragraph.gohtml", p)
	return template.HTML(bs.String()), err
}

func (cb CodeBlock) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := post.Execute(bs, "code-block.gohtml", cb)
	return template.HTML(bs.String()), err
}

func (i Image) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := post.Execute(bs, "image.gohtml", i)
	return template.HTML(bs.String()), err
}

func (v Video) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := post.Execute(bs, "video.gohtml", v)
	return template.HTML(bs.String()), err
}

func (b Blockquote) Render() (template.HTML, error) {
	bs := &bytes.Buffer{}
	err := post.Execute(bs, "blockquote.gohtml", b)
	return template.HTML(bs.String()), err
}

func (hr HorizontalRule) Render() (template.HTML, error) {
	return template.HTML("\n<hr>\n"), nil
}

func (p Post) ShowRelevantSection() bool {
	return p.Relevant != nil
}

func (r ReadingItem) FormatDate() string {
	return r.Date.Format("2006-01-02")
}

func (p Post) PublishedFull() string {
	return p.Published.Published.Format(time.RFC3339)
}

func (p Post) RevisedFull() string {
	Assert(p.Published.HasRevision(), "must check HasRevision to know if it is safe to access Revised")
	return p.Published.Revised.Format(time.RFC3339)
}

func (p Post) CopyrightYears() template.HTML {
	if p.Published.HasRevision() && (p.Published.Published.Year() != p.Published.Revised.Year()) {
		return template.HTML(fmt.Sprintf("%s&ndash;%s", p.Published.Published.Format("2006"), p.Published.Revised.Format("2006")))
	}
	return template.HTML(p.Published.Published.Format("2006"))
}

func (p Post) ObfuscatedAuthorCredit() (template.HTML, error) {
	authorName, err := p.Author.Name.Render()
	if err != nil {
		return "", err
	}
	return template.HTML(fmt.Sprintf(`<a href="mailto:%s">%s</a>`, p.ObfuscatedEmail(), authorName)), nil
}

func (p Post) ObfuscatedEmail() template.HTML {
	const (
		janetStart = `janet -e '(print (string/from-bytes (splice (map (fn [c] (if (<= 97 c 122) (+ 97 (mod (+ c -97 13) 26)) c)) &quot;`
		janetEnd   = `&quot;))))'`
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
	emailParts := strings.Split(p.Author.Email.Text(), "@")
	Assert(len(emailParts) == 2, "invalid email address") // @todo: should validate this user input in some earlier stage
	alias := emailParts[0] + fmt.Sprintf("+%s", url.QueryEscape(p.UrlPath))
	domain := emailParts[1]
	return template.HTML(janetStart + rot13(alias) + "@" + rot13(domain) + janetEnd)
}

type (
	MakeGenVisitor struct {
		//parser.NopVisitor
		TemplateData     *Post
		Errors           error
		currentSection1  *Section
		currentSection2  *Section
		currentParagraph *Paragraph
		currentSOC       StringOnlyContent

		currentContainer Container
		htmlState        HtmlState
	}
	Container interface {
		Append(r Renderable)
	}

	MakeQuotesVisitor struct {
		MakeGenVisitor
	}
)

func (s *Section) Append(r Renderable) {
	s.Content = append(s.Content, r)
}

func (v *MakeGenVisitor) VisitBlog(b *parser.Blog) {
	if draft, ok := b.Meta["draft"]; ok {
		if len(draft) > 1 {
			v.Errors = errors.Join(v.Errors, errors.New("multiple definitions of meta key: draft"))
		}
		draftVal := stringFromTextSimple(draft[0])
		if draftVal == "false" {
			v.TemplateData.MakePublish = true
		}
	}

	if urlPath, ok := b.Meta["url-path"]; ok {
		if len(urlPath) > 1 {
			v.Errors = errors.Join(v.Errors, errors.New("multiple definitions of meta key: url-path"))
		}
		v.TemplateData.UrlPath = stringFromTextSimple(urlPath[0])
	} else {
		v.Errors = errors.Join(v.Errors, errors.New("missing mandatory meta key: url-path"))
	}

	if author, ok := b.Meta["author"]; ok {
		if len(author) > 1 {
			v.Errors = errors.Join(v.Errors, errors.New("multiple definitions of meta key: author"))
		}
		v.TemplateData.Author.Name = stringRenderableFromTextSimple(author[0])
	} else {
		v.Errors = errors.Join(v.Errors, errors.New("missing mandatory meta key: author"))
	}

	if title, ok := b.Meta["title"]; ok {
		if len(title) > 1 {
			v.Errors = errors.Join(v.Errors, errors.New("multiple definitions of meta key: title"))
		}
		v.TemplateData.Title = stringRenderableFromTextSimple(title[0])
	} else {
		v.Errors = errors.Join(v.Errors, errors.New("missing mandatory meta key: title"))
	}

	if lang, ok := b.Meta["lang"]; ok {
		if len(lang) > 1 {
			v.Errors = errors.Join(v.Errors, errors.New("multiple definitions of meta key: lang"))
		}
		v.TemplateData.Lang = stringFromTextSimple(lang[0])
	} else {
		v.Errors = errors.Join(v.Errors, errors.New("missing mandatory meta key: lang"))
	}

	if email, ok := b.Meta["email"]; ok {
		if len(email) > 1 {
			v.Errors = errors.Join(v.Errors, errors.New("multiple definitions of meta key: email"))
		}
		v.TemplateData.Author.Email = stringRenderableFromTextSimple(email[0])
	}
	if relMe, ok := b.Meta["rel-me"]; ok {
		if len(relMe) > 1 {
			v.Errors = errors.Join(v.Errors, errors.New("multiple definitions of meta key: rel-me"))
		}
		v.TemplateData.Author.RelMe = stringRenderableFromTextSimple(relMe[0])
	}
	if fediCreator, ok := b.Meta["fedi-creator"]; ok {
		if len(fediCreator) > 1 {
			v.Errors = errors.Join(v.Errors, errors.New("multiple definitions of meta key: fedi-creator"))
		}
		v.TemplateData.Author.FediCreator = stringRenderableFromTextSimple(fediCreator[0])
	}
	//if template, ok := b.Meta["template"]; ok {
	//	// @todo (this probably has to be handled before / outside the visitor.
	//	// What visitor to use is decided on the template, since the visitor
	//	// has to construct a data structure specific to the template being
	//	// used.
	//}
	if description, ok := b.Meta["description"]; ok {
		if len(description) > 1 {
			v.Errors = errors.Join(v.Errors, errors.New("multiple definitions of meta key: description"))
		}
		v.TemplateData.Description = stringFromTextSimple(description[0])
	}
	if altTitle, ok := b.Meta["alt-title"]; ok {
		if len(altTitle) > 1 {
			v.Errors = errors.Join(v.Errors, errors.New("multiple definitions of meta key: alt-title"))
		}
		v.TemplateData.AltTitle = stringRenderableFromTextSimple(altTitle[0])
	}
	if published, ok := b.Meta["published"]; ok {
		if len(published) > 1 {
			v.Errors = errors.Join(v.Errors, errors.New("multiple definitions of meta key: published"))
		}
		date, err := dateFromTextSimple(published[0])
		if err != nil {
			v.Errors = errors.Join(v.Errors, fmt.Errorf("published: %w", err))
		}
		v.TemplateData.Published.Published = date
	}
	if revised, ok := b.Meta["revised"]; ok {
		if len(revised) > 1 {
			v.Errors = errors.Join(v.Errors, errors.New("multiple definitions of meta key: revised"))
		}
		date, err := dateFromTextSimple(revised[0])
		if err != nil {
			v.Errors = errors.Join(v.Errors, fmt.Errorf("revised: %w", err))
		}
		v.TemplateData.Published.Revised = &date
	}
	if estReading, ok := b.Meta["est-reading"]; ok {
		if len(estReading) > 1 {
			v.Errors = errors.Join(v.Errors, errors.New("multiple definitions of meta key: est-reading"))
		}
		i, err := intFromTextSimple(estReading[0])
		if err != nil {
			v.Errors = errors.Join(v.Errors, fmt.Errorf("est-reading: %w", err))
		}
		v.TemplateData.EstReading = i
	}
	if series, ok := b.Meta["series"]; ok {
		if len(series) > 1 {
			v.Errors = errors.Join(v.Errors, errors.New("multiple definitions of meta key: series"))
		}
		v.TemplateData.Series = &Series{
			Name: stringRenderableFromTextSimple(series[0]),
			Link: stringFromTextSimple(series[0]), // @todo: allow custom link
		}
	}
	if enableRevisionWarning, ok := b.Meta["enable-revision-warning"]; ok {
		if len(enableRevisionWarning) > 1 {
			v.Errors = errors.Join(v.Errors, errors.New("multiple definitions of meta key: enable-revision-warning"))
		}
		b, err := boolFromTextSimple(enableRevisionWarning[0])
		if err != nil {
			v.Errors = errors.Join(v.Errors, fmt.Errorf("enable-revision-warning: %w", err))
		}
		v.TemplateData.EnableRevisionWarning = b
	}
	if tags, ok := b.Meta["tags"]; ok {
		var tagStrs []string
		for _, t := range tags {
			tagStrs = append(tagStrs, strings.Split(stringFromTextSimple(t), " ")...)
		}
		for _, t := range tagStrs {
			v.TemplateData.Tags = append(v.TemplateData.Tags, Tag(t))
		}
	}
}

func (v *MakeGenVisitor) VisitSection(s *parser.Section) {
	switch s.Level {
	default:
		panic(fmt.Errorf("invalid section level: %d", s.Level))
	case 1:
		v.currentSection1 = &Section{
			Attributes: Attributes(s.Attributes),
			Level:      s.Level,
			Heading:    stringRenderableFromTextRich(s.Heading),
		}
		v.currentContainer = v.currentSection1
		v.TemplateData.TOC.Sections = append(v.TemplateData.TOC.Sections, TOCSection{
			ID:      v.currentSection1.ID(),
			Heading: v.currentSection1.Heading,
		})
	case 2:
		v.currentSection2 = &Section{
			Attributes: Attributes(s.Attributes),
			Level:      s.Level,
			Heading:    stringRenderableFromTextRich(s.Heading),
		}
		v.currentContainer = v.currentSection2
		{
			l := len(v.TemplateData.TOC.Sections)
			v.TemplateData.TOC.Sections[l-1].NextLevel = append(v.TemplateData.TOC.Sections[l-1].NextLevel, TOCSection{
				ID:      v.currentSection2.ID(),
				Heading: v.currentSection2.Heading,
			})
		}
	}
}

func (v *MakeGenVisitor) VisitParagraph(p *parser.Paragraph) {
	v.currentParagraph = &Paragraph{}
}

func (v *MakeGenVisitor) VisitText(t *parser.Text) {
	v.currentSOC = append(v.currentSOC, Text(*t))
}

func (v *MakeGenVisitor) VisitLink(l *parser.Link) {
	v.currentSOC = append(v.currentSOC, Link{
		Name: stringRenderableFromTextRich(l.Name),
		Href: l.Href,
	})
}

func (v *MakeGenVisitor) VisitSidenote(s *parser.Sidenote) {
	v.currentSOC = append(v.currentSOC, Sidenote{
		Word:    stringRenderableFromTextRich(s.Word),
		Content: stringRenderableFromTextRich(s.Content),
	})
}

func (v *MakeGenVisitor) VisitAmpSpecial(a *parser.AmpSpecial) {
	v.currentSOC = append(v.currentSOC, getAmpSpecial(string(*a)))
}

func (v *MakeGenVisitor) VisitEmphasis(e *parser.Emphasis) {
	v.currentSOC = append(v.currentSOC, Emphasis{stringRenderableFromTextRich(parser.TextRich(*e))})
}

func (v *MakeGenVisitor) VisitStrong(e *parser.Strong) {
	v.currentSOC = append(v.currentSOC, Strong{stringRenderableFromTextRich(parser.TextRich(*e))})
}

func (v *MakeGenVisitor) VisitEmphasisStrong(e *parser.EmphasisStrong) {
	v.currentSOC = append(v.currentSOC, EmphasisStrong{stringRenderableFromTextRich(parser.TextRich(*e))})
}

func (v *MakeGenVisitor) VisitEnquoteDouble(e *parser.EnquoteDouble) {
	v.currentSOC = append(v.currentSOC, EnquoteDouble{stringRenderableFromTextRich(parser.TextRich(*e))})
}

func (v *MakeGenVisitor) VisitEnquoteAngled(e *parser.EnquoteAngled) {
	v.currentSOC = append(v.currentSOC, EnquoteAngled{stringRenderableFromTextRich(parser.TextRich(*e))})
}

func (v *MakeGenVisitor) VisitLinkify(l *parser.Linkify) {
	v.currentSOC = append(v.currentSOC, Link{
		Name: StringOnlyContent{Text(*l)},
		Href: string(*l),
	})
}

func (v *MakeGenVisitor) VisitMarker(m *parser.Marker) {
	v.currentSOC = append(v.currentSOC, Marker{
		stringRenderableFromTextRich(parser.TextRich(*m)),
	})
}

func (v *MakeGenVisitor) VisitMono(m *parser.Mono) {
	v.currentSOC = append(v.currentSOC, Mono(*m))
}

func (v *MakeGenVisitor) VisitStrikethrough(s *parser.Strikethrough) {
	v.currentSOC = append(v.currentSOC, Strikethrough{
		stringRenderableFromTextRich(parser.TextRich(*s)),
	})
}

func (v *MakeGenVisitor) VisitLineBreak(h *parser.LineBreak) {
	v.currentSOC = append(v.currentSOC, LineBreak{})
}

func (v *MakeGenVisitor) LeaveParagraph(p *parser.Paragraph) {
	v.currentParagraph.Content = v.currentSOC
	v.currentSOC = nil
	v.currentContainer.Append(*v.currentParagraph)
	v.currentParagraph = nil
}

func (v *MakeGenVisitor) VisitBlockQuote(b *parser.BlockQuote) {
	v.currentContainer.Append(Blockquote{
		QuoteText: stringRenderableFromTextRich(b.QuoteText),
		Author:    stringRenderableFromTextSimple(b.Author),
		Source:    stringRenderableFromTextRich(b.Source),
	})
}

func (v *MakeGenVisitor) VisitCodeBlock(c *parser.CodeBlock) {
	v.currentContainer.Append(CodeBlock{
		Attributes: Attributes(c.Attributes),
		Lines:      c.Lines,
	})
}

func (v *MakeGenVisitor) VisitHorizontalRule(h *parser.HorizontalRule) {
	v.currentContainer.Append(HorizontalRule{})
}

// @todo: bad name, since it can also be a video now
func (v *MakeGenVisitor) VisitImage(i *parser.Image) {
	// @todo: we need to keep this synchronized with the one in markup.go
	switch filepath.Ext(i.Name) {
	default:
		panic("unreachable")
	case ".jpg", ".jpeg", ".jxl", ".avif", ".webp", ".png":
		v.currentContainer.Append(Image{
			Name:  strings.TrimSuffix(filepath.Base(i.Name), filepath.Ext(i.Name)),
			Alt:   stringRenderableFromTextSimple(i.Alt),
			Title: stringRenderableFromTextSimple(i.Title),
		})
	case ".mp4", ".mkv", ".webm":
		v.currentContainer.Append(Video{
			Name:  strings.TrimSuffix(filepath.Base(i.Name), filepath.Ext(i.Name)),
			Alt:   stringRenderableFromTextSimple(i.Alt),
			Title: stringRenderableFromTextSimple(i.Title),
		})
	}
}

func (v *MakeGenVisitor) LeaveSection(s *parser.Section) {
	switch s.Level {
	default:
		panic(fmt.Errorf("invalid section level: %d", s.Level))
	case 1:
		v.TemplateData.Sections = append(v.TemplateData.Sections, *v.currentSection1)
		v.currentSection1 = nil
		v.currentContainer = nil
	case 2:
		v.currentSection1.Content = append(v.currentSection1.Content, *v.currentSection2)
		v.currentSection2 = nil
		v.currentContainer = v.currentSection1
	}
}

func (v *MakeGenVisitor) VisitHtml(h *parser.Html) {
	if v.htmlState == nil {
		v.htmlState = v.htmlTopLevel
	}
	if v.currentContainer == nil {
		v.currentContainer = HtmlTopLevel{Gen: v}
	}
	v.htmlState(v, h, true)
}

func (v *MakeGenVisitor) LeaveHtml(h *parser.Html) {
	v.htmlState(v, h, false)
}

var ErrInvalidHtmlPos = errors.New("html element not allowed in this position")

type (
	HtmlState   func(v *MakeGenVisitor, h *parser.Html, entering bool)
	HtmlInvalid struct {
		nestingCount int
	}
	HtmlAbstract struct {
		content []Renderable
		err     error
	}
	HtmlRelevantBox struct {
		relevantBox *RelevantBox
		currentItem *ReadingItem
		err         error
	}
	HtmlRuby struct {
		parentSOC StringOnlyContent
		furi      StringRenderable
		err       error
	}
	HtmlNote struct {
		parentContainer Container
		noteItem  Note
		err       error
	}
	HtmlTopLevel struct {
		Gen *MakeGenVisitor
	}
)

func (h HtmlTopLevel) Append(r Renderable) {
	h.Gen.TemplateData.TopLevelContent = append(h.Gen.TemplateData.TopLevelContent, r)
}

func (h *HtmlInvalid) Append(r Renderable) {
	// ignore
}

func (h *HtmlAbstract) Append(r Renderable) {
	// @todo: maybe don't allow just anything here... (only paragraphs, no sections...) [:abstract-content:]
	h.content = append(h.content, r)
}

func (h *HtmlRelevantBox) Append(r Renderable) {
	if h.currentItem == nil {
		h.err = errors.Join(h.err, fmt.Errorf("<RelevantBox> cannot contain content other than <Relevant>: %v", r))
	} else {
		// @todo: maybe don't allow just anything here... (only paragraphs, no sections...) [:abstract-content:]
		h.currentItem.Abstract = append(h.currentItem.Abstract, r)
	}
}

func (h *HtmlNote) Append(r Renderable) {
	h.noteItem.Content = append(h.noteItem.Content, r)
}

func (i *HtmlInvalid) htmlInvalid(v *MakeGenVisitor, h *parser.Html, entering bool) {
	if entering {
		i.nestingCount++
	} else {
		i.nestingCount--
		if i.nestingCount == 0 {
			v.htmlState = v.htmlTopLevel
		}
	}
}

func (r *HtmlRelevantBox) htmlRelevantBox(v *MakeGenVisitor, h *parser.Html, entering bool) {
	if entering {
		switch h.Name {
		default:
			v.Errors = errors.Join(v.Errors, fmt.Errorf("%s: %w", h.Name, ErrInvalidHtmlPos))
		case "Relevant":
			href, ok := h.Attributes["href"]
			if !ok {
				v.Errors = errors.Join(v.Errors, errors.New("relevant item missing its href attribute"))
			}
			date, ok := h.Attributes["date"]
			var parsedDate time.Time
			if !ok {
				v.Errors = errors.Join(v.Errors, errors.New("relevant item missing its date attribute"))
			} else {
				var err error
				parsedDate, err = time.Parse("2006-01-02", date)
				if err != nil {
					v.Errors = errors.Join(v.Errors, fmt.Errorf("relevant item invalid value for date attribute: %w", err))
				}
			}
			title, ok := h.Attributes["title"]
			var parsedTitle StringRenderable
			if !ok {
				v.Errors = errors.Join(v.Errors, errors.New("relevant item missing its title attribute"))
			} else {
				p, err := parseStringAsTextRich(title)
				if err != nil {
					v.Errors = errors.Join(v.Errors, fmt.Errorf("invalid value for title: %w", err))
				} else {
					parsedTitle = stringRenderableFromTextRich(p)
				}
			}
			r.currentItem = &ReadingItem{
				Link:  href,
				Title: parsedTitle,
				Date:  parsedDate,
			}
			v.htmlState = r.htmlRelevantItem
		}
	} else {
		v.TemplateData.Relevant = r.relevantBox
		v.Errors = errors.Join(v.Errors, r.err)
		v.htmlState = v.htmlTopLevel
	}
}

func (r *HtmlRelevantBox) htmlRelevantItem(v *MakeGenVisitor, h *parser.Html, entering bool) {
	if entering {
		switch h.Name {
		default:
		case "Author":
			href := h.Attributes["href"] // optional
			name, ok := h.Attributes["name"]
			var parsedName StringRenderable
			if !ok {
				v.Errors = errors.Join(v.Errors, errors.New("relevant item's author is missing its name attribute"))
			} else {
				p, err := parseStringAsTextRich(name)
				if err != nil {
					v.Errors = errors.Join(v.Errors, fmt.Errorf("invalid value for name: %w", err))
				} else {
					parsedName = stringRenderableFromTextRich(p)
				}
			}
			r.currentItem.AuthorLink = href
			r.currentItem.Author = parsedName
			v.htmlState = r.htmlRelevantItemAuthor
		case "Abstract":
			v.htmlState = r.htmlRelevantItemAbstract
		}
	} else {
		r.relevantBox.Articles = append(r.relevantBox.Articles, *r.currentItem)
		r.currentItem = nil
		v.htmlState = r.htmlRelevantBox
	}
}

func (r *HtmlRelevantBox) htmlRelevantItemAuthor(v *MakeGenVisitor, h *parser.Html, entering bool) {
	if entering {
		// invalid
		v.Errors = errors.Join(v.Errors, fmt.Errorf("<RelevantItem><Author> cannot contain any content: %s", h.Name))
	} else {
		v.htmlState = r.htmlRelevantItem
	}
}

func (r *HtmlRelevantBox) htmlRelevantItemAbstract(v *MakeGenVisitor, h *parser.Html, entering bool) {
	if entering {
		// invalid
		v.Errors = errors.Join(v.Errors, fmt.Errorf("<RelevantItem><Abstract> cannot contain any child html elements: %s", h.Name))
	} else {
		v.htmlState = r.htmlRelevantItem
	}
}

func (v *MakeGenVisitor) htmlTopLevel(_ *MakeGenVisitor, h *parser.Html, entering bool) {
	if entering {
		switch h.Name {
		default:
			v.Errors = errors.Join(v.Errors, fmt.Errorf("%s: %w", h.Name, ErrInvalidHtmlPos))
			i := &HtmlInvalid{nestingCount: 1}
			v.currentContainer = i
			v.htmlState = i.htmlInvalid
		case "Abstract":
			a := &HtmlAbstract{}
			v.currentContainer = a
			v.htmlState = a.htmlAbstract
		case "RelevantBox":
			heading := StringOnlyContent{Text("Articles from blogs I read")}
			if customHeading, ok := h.Attributes["title"]; ok {
				p, err := parseStringAsTextRich(customHeading)
				if err != nil {
					v.Errors = errors.Join(v.Errors, fmt.Errorf("invalid value for heading: %w", err))
				} else {
					heading = stringRenderableFromTextRich(p)
				}
			}
			r := &HtmlRelevantBox{
				relevantBox: &RelevantBox{
					Heading: heading,
				},
			}
			v.currentContainer = r
			v.htmlState = r.htmlRelevantBox
		case "Note":
			noteType := h.Attributes["type"]
			n := &HtmlNote{
				parentContainer: v.currentContainer,
				noteItem: Note{
					Type: noteType,
				},
			}
			v.currentContainer = n
			v.htmlState = n.htmlNote
		case "Ruby":
			var furi StringRenderable
			if furiAttr, ok := h.Attributes["furi"]; ok {
				furiRich, err := parseStringAsTextRich(furiAttr)
				if err != nil {
					v.Errors = errors.Join(v.Errors, fmt.Errorf("invalid value for furi: %w", err))
				} else {
					furi = stringRenderableFromTextRich(furiRich)
				}
			} else {
				v.Errors = errors.Join(v.Errors, errors.New("ruby element missing its furi attribute"))
			}
			r := &HtmlRuby{
				parentSOC: v.currentSOC,
				furi:      furi,
			}
			v.currentSOC = nil
			v.htmlState = r.htmlRuby
		}
	} else {
		// invalid
		panic("html state confused")
	}
}

func (a *HtmlAbstract) htmlAbstract(v *MakeGenVisitor, h *parser.Html, entering bool) {
	if entering {
		// invalid
		v.Errors = errors.Join(v.Errors, fmt.Errorf("<Abstract> cannot contain any child html elements: %s", h.Name))
	} else {
		v.currentContainer = nil
		v.Errors = errors.Join(v.Errors, a.err)
		v.TemplateData.Abstract = a.content
		v.htmlState = v.htmlTopLevel
	}
}

func (n *HtmlNote) htmlNote(v *MakeGenVisitor, h *parser.Html, entering bool) {
	if entering {
		v.Errors = errors.Join(v.Errors, fmt.Errorf("<Ruby> cannot contain any child html elements: %s", h.Name))
	} else {
		v.Errors = errors.Join(v.Errors, n.err)
		n.parentContainer.Append(n.noteItem)
		v.currentContainer = n.parentContainer
		v.htmlState = v.htmlTopLevel
	}
}

func (r *HtmlRuby) htmlRuby(v *MakeGenVisitor, h *parser.Html, entering bool) {
	if entering {
		v.Errors = errors.Join(v.Errors, fmt.Errorf("<Ruby> cannot contain any child html elements: %s", h.Name))
	} else {
		v.Errors = errors.Join(v.Errors, r.err)
		r.parentSOC = append(r.parentSOC, Ruby{
			Kanji:    v.currentSOC,
			Furigana: r.furi,
		})
		v.currentSOC = r.parentSOC
		v.htmlState = v.htmlTopLevel
	}
}

func (v *MakeGenVisitor) LeaveBlog(b *parser.Blog) {
}

func (v *MakeQuotesVisitor) VisitBlockQuote(b *parser.BlockQuote) {
	text := stringRenderableFromTextRich(b.QuoteText)
	author := stringRenderableFromTextSimple(b.Author)
	source := stringRenderableFromTextRich(b.Source)
	hashID := sha256.New()
	hashID.Write([]byte(text.Text()))
	hashID.Write([]byte(author.Text()))
	hashID.Write([]byte(source.Text()))
	v.currentContainer.Append(Blockquote{
		Attributes: Attributes{
			"id":    base64.StdEncoding.EncodeToString(hashID.Sum(nil)),
			"class": "quote",
		},
		QuoteText: text,
		Author:    author,
		Source:    source,
	})
}

func stringFromTextSimple(t parser.TextSimple) string {
	var b strings.Builder
	for _, n := range t {
		switch e := n.(type) {
		default:
			panic(fmt.Errorf("%T cannot be converted to string", e))
		case *parser.Text:
			b.WriteString(string(*e))
		case *parser.AmpSpecial:
			b.WriteString(string(*e))
		}
	}
	return b.String()
}

func stringRenderableFromTextSimple(t parser.TextSimple) StringOnlyContent {
	var soc StringOnlyContent
	for _, n := range t {
		switch e := n.(type) {
		default:
			panic(fmt.Errorf("%T cannot be converted to StringRenderable", e))
		case *parser.Text:
			soc = append(soc, Text(*e))
		case *parser.AmpSpecial:
			soc = append(soc, getAmpSpecial(string(*e)))
		}
	}
	return soc
}

func stringRenderableFromTextRich(t parser.TextRich) StringOnlyContent {
	var soc StringOnlyContent
	for _, n := range t {
		switch e := n.(type) {
		default:
			panic(fmt.Errorf("%T cannot be converted to StringRenderable", e))
		case *parser.Text:
			soc = append(soc, Text(*e))
		case *parser.AmpSpecial:
			soc = append(soc, getAmpSpecial(string(*e)))
		case *parser.Emphasis:
			soc = append(soc, Emphasis{stringRenderableFromTextRich(parser.TextRich(*e))})
		case *parser.Strong:
			soc = append(soc, Strong{stringRenderableFromTextRich(parser.TextRich(*e))})
		case *parser.EmphasisStrong:
			soc = append(soc, EmphasisStrong{stringRenderableFromTextRich(parser.TextRich(*e))})
		case *parser.Link:
			soc = append(soc, Link{
				Name: stringRenderableFromTextRich(e.Name),
				Href: e.Href,
			})
		case *parser.Sidenote:
			soc = append(soc, Sidenote{
				Word:    stringRenderableFromTextRich(e.Word),
				Content: stringRenderableFromTextRich(e.Content),
			})
		case *parser.Strikethrough:
			soc = append(soc, Strikethrough{stringRenderableFromTextRich(parser.TextRich(*e))})
		case *parser.Marker:
			soc = append(soc, Marker{stringRenderableFromTextRich(parser.TextRich(*e))})
		case *parser.Mono:
			soc = append(soc, Mono(*e))
		case *parser.Linkify:
			soc = append(soc, Link{
				Href: string(*e),
			})
		case *parser.EnquoteDouble:
			soc = append(soc, EnquoteDouble{stringRenderableFromTextRich(parser.TextRich(*e))})
		case *parser.EnquoteAngled:
			soc = append(soc, EnquoteAngled{stringRenderableFromTextRich(parser.TextRich(*e))})
		case *parser.LineBreak:
			soc = append(soc, LineBreak{})
		}
	}
	return soc
}

func getAmpSpecial(s string) EscapedString {
	switch s {
	default:
		panic(fmt.Errorf("not an amp special: %s", s))
	case "~", "\u00A0", "&nbsp;":
		return AmpNoBreakSpace
	case "--", "&ndash;":
		return AmpEnDash
	case "---", "mdash;":
		return AmpEmDash
	case "&ldquo;":
		return AmpLeftDoubleQuote
	case "&rdquo;":
		return AmpRightDoubleQuote
	case "...", "…":
		return AmpEllipsis
	case "&prime;":
		return AmpPrime
	case "&Prime;":
		return AmpDoublePrime
	case "&tprime;":
		return AmpTripplePrime
	case "&qprime;":
		return AmpQuadruplePrime
	case "&bprime;":
		return AmpReversedPrime
	case "&laquo;":
		return AmpLeftAngledQuote
	case "&raquo;":
		return AmpRightAngledQuote
	case "&rsquo;", "'":
		return AmpRightSingleQuoteOrApostrophe
	case "&lsquo;":
		return AmpLeftSingleQuote
	}
}

func dateFromTextSimple(t parser.TextSimple) (date time.Time, err error) {
	dateStr, ok := (t[0]).(*parser.Text)
	if !ok {
		return date, fmt.Errorf("cannot convert to date: %s", t)
	}
	date, err = time.Parse(time.RFC3339, string(*dateStr))
	if err != nil {
		date, err = time.Parse("2006-01-02", string(*dateStr))
	}
	return date, err
}

func intFromTextSimple(t parser.TextSimple) (int, error) {
	intStr, ok := (t[0]).(*parser.Text)
	if !ok {
		return 0, fmt.Errorf("cannot convert to int: %s", t)
	}
	i, err := strconv.Atoi(string(*intStr))
	return i, err
}

func boolFromTextSimple(t parser.TextSimple) (bool, error) {
	boolStr, ok := (t[0]).(*parser.Text)
	if !ok {
		return false, fmt.Errorf("cannot convert to bool: %s", t)
	}
	switch string(*boolStr) {
	default:
		return false, fmt.Errorf("not a boolean: %v", boolStr)
	case "false":
		return false, nil
	case "true":
		return true, nil
	}
}

func parseStringAsTextRich(s string) (parser.TextRich, error) {
	// @todo:
	return parser.TextRich{AsRef(parser.Text(s))}, nil
}
