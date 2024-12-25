package gen

import (
	"errors"
	"time"
	"fmt"
	"strconv"
	"strings"

	"github.com/cvanloo/blog-go/markup/parser"
)

type (
	MakeGenVisitor struct {
		parser.NopVisitor
		TemplateData *Blog
		Errors       error
		currentSection1 *Section
		currentSection2 *Section
		currentSection *Section
		currentParagraph *Paragraph
		currentSOC StringOnlyContent
	}
)

func (v *MakeGenVisitor) VisitBlog(b *parser.Blog) {
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
		// @todo
		_ = series
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
			Level: s.Level,
			Heading: stringRenderableFromTextRich(s.Heading),
		}
		v.currentSection = v.currentSection1
		v.TemplateData.TOC.Sections = append(v.TemplateData.TOC.Sections, TOCSection{
			ID: v.currentSection.ID(),
			Heading: v.currentSection.Heading,
		})
	case 2:
		v.currentSection2 = &Section{
			Attributes: Attributes(s.Attributes),
			Level: s.Level,
			Heading: stringRenderableFromTextRich(s.Heading),
		}
		v.currentSection = v.currentSection2
		{
			l := len(v.TemplateData.TOC.Sections)
			v.TemplateData.TOC.Sections[l-1].NextLevel = append(v.TemplateData.TOC.Sections[l-1].NextLevel, TOCSection{
				ID: v.currentSection.ID(),
				Heading: v.currentSection.Heading,
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
		Word: stringRenderableFromTextRich(s.Word),
		Content: stringRenderableFromTextRich(s.Content),
	})
}

func (v *MakeGenVisitor) VisitAmpSpecial(a *parser.AmpSpecial) {
	v.currentSOC = append(v.currentSOC, getAmpSpecial(string(*a)))
}

func (v *MakeGenVisitor) LeaveParagraph(p *parser.Paragraph) {
	v.currentParagraph.Content = v.currentSOC
	v.currentSOC = nil
	v.currentSection.Content = append(v.currentSection.Content, *v.currentParagraph)
	v.currentParagraph = nil
}

func (v *MakeGenVisitor) LeaveSection(s *parser.Section) {
	switch s.Level {
	default:
		panic(fmt.Errorf("invalid section level: %d", s.Level))
	case 1:
		v.TemplateData.Sections = append(v.TemplateData.Sections, *v.currentSection1)
		v.currentSection1 = nil
		v.currentSection = nil
	case 2:
		v.currentSection1.Content = append(v.currentSection1.Content, *v.currentSection2)
		v.currentSection2 = nil
		v.currentSection = v.currentSection1
	}
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
				Word: stringRenderableFromTextRich(e.Word),
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
