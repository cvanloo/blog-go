package parser

import (
	"errors"
	"fmt"
	"strings"

	//"github.com/kr/pretty"

	. "github.com/cvanloo/blog-go/assert"
	"github.com/cvanloo/blog-go/markup/lexer"
)

/*
type (
	// NonHtmlText cannot contain HTML elements like <strong>, but allows for
	// amp specials such as &nbsp;
	NonHtmlText interface {
		Text() string
	}

	// ParagraphContent is any (text) element that can appear only inside a paragraph.
	// Examples include <strong>, sidenotes, enquotes, any NonHtmlText.
	ParagraphContent interface {
		NonHtmlText
		RichText() template.HTML
	}

	// SectionContent is any element that can appear only inside a section.
	// Examples include section 2, images, blockquotes.
	SectionContent interface {
		//ParagraphContent
		Render() template.HTML
	}
)
*/

type (
	TextSimple []Node
	TextRich   []Node

	Node interface {
		Accept(Visitor)
	}
	Visitor interface {
		VisitBlog(*Blog)
		LeaveBlog(*Blog)
		VisitSection(*Section)
		LeaveSection(*Section)
		VisitParagraph(*Paragraph)
		LeaveParagraph(*Paragraph)
		VisitLink(*Link)
		VisitSidenote(*Sidenote)
		VisitEnquoteDouble(*EnquoteDouble)
		VisitEnquoteAngled(*EnquoteAngled)
		VisitEmphasis(*Emphasis)
		VisitStrong(*Strong)
		VisitEmphasisStrong(*EmphasisStrong)
		VisitStrikethrough(*Strikethrough)
		VisitMarker(*Marker)
		VisitMono(*Mono)
		VisitText(*Text)
		VisitAmpSpecial(*AmpSpecial)
		VisitLinkify(*Linkify)
		VisitImage(*Image)
		VisitBlockQuote(*BlockQuote)
		VisitCodeBlock(*CodeBlock)
		VisitHorizontalRule(*HorizontalRule)
		VisitLineBreak(*LineBreak)
		VisitHtml(*Html)
		LeaveHtml(*Html)
	}

	Attributes map[string]string
	Section    struct {
		Attributes
		Level   int
		Heading TextRich
		Content []Node
	}
	Paragraph struct {
		Content []Node
	}
	Link struct {
		Ref  string
		Name TextRich
		Href string
	}
	Sidenote struct {
		Ref           string
		Word, Content TextRich
	}
	Image struct {
		Name       string
		Alt, Title TextSimple
	}
	BlockQuote struct {
		QuoteText TextRich
		Author    TextSimple
		Source    TextRich
	}
	CodeBlock struct {
		Attributes
		Lines []string
	}
	HorizontalRule struct{}
	LineBreak      struct{}
	EnquoteDouble  TextRich
	EnquoteAngled  TextRich
	Emphasis       TextRich
	Strong         TextRich
	EmphasisStrong TextRich
	Strikethrough  TextRich
	Marker         TextRich
	Mono           string
	Text           string
	AmpSpecial     string
	Linkify        string
	Html           struct {
		Attributes
		Name    string
		Content []Node
	}

	NopVisitor           struct{}
	FixReferencesVisitor struct {
		NopVisitor
		Errors              error
		LinkDefinitions     map[string]string
		SidenoteDefinitions map[string]TextRich
		TermDefinitions     map[string]TextRich
	}

	Blog struct {
		Meta                Meta
		Sections            []*Section
		Htmls               []*Html
		LinkDefinitions     map[string]string
		SidenoteDefinitions map[string]TextRich
		TermDefinitions     map[string]TextRich
	}
	Meta map[string][]TextSimple // slice value to allow for duplicate keys
)

func getText(t TextSimple) (string, bool) {
	var builder strings.Builder
	for _, n := range t {
		if e, ok := n.(*Text); ok {
			builder.WriteString(string(*e))
		} else {
			return builder.String(), false
		}
	}
	return builder.String(), true
}

func (m Meta) Template() (string, bool) {
	ts, ok := m["template"]
	if !ok {
		return "", false
	}
	if len(ts) <= 0 {
		return "", false
	}
	return getText(ts[0])
}

func (t *TextSimple) Append(n Node) bool {
	switch n.(type) {
	default:
		return false
	case *Text, *AmpSpecial:
		*t = append(*t, n)
		return true
	}
}

func (t *TextRich) Append(n Node) bool {
	switch n.(type) {
	default:
		return false
	case *Text, *AmpSpecial, *Emphasis, *Strong, *EmphasisStrong, *Link, *Sidenote, *Strikethrough, *Marker, *Mono, *Linkify, *EnquoteDouble, *EnquoteAngled, *LineBreak:
		*t = append(*t, n)
		return true
	}
}

func isTextNode(token lexer.Token) bool {
	switch token.Type {
	default:
		return false
	case lexer.TokenMono, lexer.TokenText, lexer.TokenAmpSpecial, lexer.TokenLinkify, lexer.TokenLineBreak, lexer.TokenImageAltText, lexer.TokenImageTitle:
		// @todo: what else is a text node?
		return true
	}
	panic("unreachable")
}

func newTextNode(lexeme lexer.Token) Node {
	Assert(isTextNode(lexeme), fmt.Sprintf("cannot make text node out of %s", lexeme.Type))
	switch lexeme.Type {
	case lexer.TokenMono:
		return AsRef(Mono(lexeme.Text))
	case lexer.TokenText:
		return AsRef(Text(lexeme.Text))
	case lexer.TokenLinkify:
		return AsRef(Linkify(lexeme.Text))
	case lexer.TokenAmpSpecial:
		return AsRef(AmpSpecial(lexeme.Text))
	case lexer.TokenLineBreak:
		return AsRef(LineBreak{})
	case lexer.TokenImageAltText:
		// @todo: really? ^
		return AsRef(Text(lexeme.Text))
	case lexer.TokenImageTitle:
		// @todo: really? ^
		return AsRef(Text(lexeme.Text))
	}
	panic("unreachable")
}

func (b *Blog) Accept(v Visitor) {
	v.VisitBlog(b)
	for _, h := range b.Htmls {
		h.Accept(v)
	}
	for _, s := range b.Sections {
		s.Accept(v)
	}
	v.LeaveBlog(b)
}

func (s *Section) Accept(v Visitor) {
	v.VisitSection(s)
	for _, c := range s.Content {
		c.Accept(v)
	}
	v.LeaveSection(s)
}

func (p *Paragraph) Accept(v Visitor) {
	v.VisitParagraph(p)
	for _, c := range p.Content {
		c.Accept(v)
	}
	v.LeaveParagraph(p)
}

func (l *Link) Accept(v Visitor) {
	v.VisitLink(l)
}

func (s *Sidenote) Accept(v Visitor) {
	v.VisitSidenote(s)
}

func (e *EnquoteDouble) Accept(v Visitor) {
	v.VisitEnquoteDouble(e)
}

func (e *EnquoteAngled) Accept(v Visitor) {
	v.VisitEnquoteAngled(e)
}

func (e *Emphasis) Accept(v Visitor) {
	v.VisitEmphasis(e)
}

func (s *Strong) Accept(v Visitor) {
	v.VisitStrong(s)
}

func (e *EmphasisStrong) Accept(v Visitor) {
	v.VisitEmphasisStrong(e)
}

func (s *Strikethrough) Accept(v Visitor) {
	v.VisitStrikethrough(s)
}

func (m *Marker) Accept(v Visitor) {
	v.VisitMarker(m)
}

func (m *Mono) Accept(v Visitor) {
	v.VisitMono(m)
}

func (t *Text) Accept(v Visitor) {
	v.VisitText(t)
}

func (a *AmpSpecial) Accept(v Visitor) {
	v.VisitAmpSpecial(a)
}

func (l *Linkify) Accept(v Visitor) {
	v.VisitLinkify(l)
}

func (i *Image) Accept(v Visitor) {
	v.VisitImage(i)
}

func (b *BlockQuote) Accept(v Visitor) {
	v.VisitBlockQuote(b)
}

func (c *CodeBlock) Accept(v Visitor) {
	v.VisitCodeBlock(c)
}

func (h *HorizontalRule) Accept(v Visitor) {
	v.VisitHorizontalRule(h)
}

func (h *LineBreak) Accept(v Visitor) {
	v.VisitLineBreak(h)
}

func (h *Html) Accept(v Visitor) {
	v.VisitHtml(h)
	for _, c := range h.Content {
		c.Accept(v)
	}
	v.LeaveHtml(h)
}

func (v NopVisitor) VisitBlog(b *Blog) {
}

func (v NopVisitor) LeaveBlog(b *Blog) {
}

func (v NopVisitor) VisitSection(*Section) {
}

func (v NopVisitor) LeaveSection(*Section) {
}

func (v NopVisitor) VisitParagraph(*Paragraph) {
}

func (v NopVisitor) LeaveParagraph(*Paragraph) {
}

func (v NopVisitor) VisitLink(*Link) {
}

func (v NopVisitor) VisitSidenote(*Sidenote) {
}

func (v NopVisitor) VisitImage(*Image) {
}

func (v NopVisitor) VisitBlockQuote(*BlockQuote) {
}

func (v NopVisitor) VisitEnquoteDouble(*EnquoteDouble) {
}

func (v NopVisitor) VisitEnquoteAngled(*EnquoteAngled) {
}

func (v NopVisitor) VisitEmphasis(*Emphasis) {
}

func (v NopVisitor) VisitStrong(*Strong) {
}

func (v NopVisitor) VisitEmphasisStrong(*EmphasisStrong) {
}

func (v NopVisitor) VisitStrikethrough(*Strikethrough) {
}

func (v NopVisitor) VisitMarker(*Marker) {
}

func (v NopVisitor) VisitMono(*Mono) {
}

func (v NopVisitor) VisitText(*Text) {
}

func (v NopVisitor) VisitAmpSpecial(*AmpSpecial) {
}

func (v NopVisitor) VisitLinkify(*Linkify) {
}

func (v NopVisitor) VisitCodeBlock(*CodeBlock) {
}

func (v NopVisitor) VisitHorizontalRule(*HorizontalRule) {
}

func (v NopVisitor) VisitLineBreak(*LineBreak) {
}

func (v NopVisitor) VisitHtml(*Html) {
}

func (v NopVisitor) LeaveHtml(*Html) {
}

func (v *FixReferencesVisitor) VisitBlog(b *Blog) {
	v.LinkDefinitions = b.LinkDefinitions
	v.SidenoteDefinitions = b.SidenoteDefinitions
	v.TermDefinitions = b.TermDefinitions
}

func (v *FixReferencesVisitor) VisitLink(l *Link) {
	if len(l.Href) == 0 {
		href, hasHref := v.LinkDefinitions[l.Ref]
		if hasHref {
			l.Href = href
		} else {
			v.Errors = errors.Join(v.Errors, fmt.Errorf("missing url definition for link with id: %s", l.Ref))
		}
	}
}

func (v *FixReferencesVisitor) VisitSidenote(sn *Sidenote) {
	content, hasContent := v.SidenoteDefinitions[sn.Ref]
	if hasContent {
		sn.Content = content
	} else {
		v.Errors = errors.Join(v.Errors, fmt.Errorf("missing content definition for sidenote with id: %s", sn.Ref))
	}
}

type (
	LexResult interface {
		Tokens() func(func(lexer.Token) bool)
	}
	ParserError struct {
		State ParseState
		Token lexer.Token
		Inner error
	}
)

// @todo: make parser a type like with the lexer?
func newError(lexeme lexer.Token, state ParseState, inner error) error {
	if inner == nil {
		return nil
	}
	return ParserError{
		State: state,
		Token: lexeme,
		Inner: inner,
	}
}

func (err ParserError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", err.State, err.Token, err.Inner)
}

type (
	Level struct {
		ReturnToState ParseState
		Strings       []string
		TextSimple    TextSimple
		TextRich      TextRich
		Content       []Node
		Html          *Html
	}
	Levels struct {
		levels []*Level
	}
)

func (ls *Levels) Push(l *Level) {
	ls.levels = append(ls.levels, l)
}

func (ls *Levels) Top() *Level {
	l := len(ls.levels)
	return ls.levels[l-1]
}

func (ls *Levels) Pop() *Level {
	l := len(ls.levels)
	top := ls.levels[l-1]
	ls.levels = ls.levels[:l-1]
	return top
}

func (ls *Levels) Len() int {
	return len(ls.levels)
}

func (l *Level) PushString(s string) {
	l.Strings = append(l.Strings, s)
}

func (l *Level) PopString() (s string) {
	s = l.Strings[len(l.Strings)-1]
	l.Strings = l.Strings[:len(l.Strings)-1]
	return s
}

// @todo: really stupid function, remove it
func (l *Level) Clear() {
	l.Strings = []string{}
	l.TextSimple = TextSimple{}
	l.TextRich = TextRich{}
	l.Content = []Node{}
}

//go:generate stringer -type ParseState
type ParseState int

const (
	ParsingStart ParseState = iota
	ParsingDocument
	ParsingMeta
	ParsingMetaVal
	ParsingHtmlElement
	ParsingHtmlElementAttributes
	ParsingHtmlElementContent
	ParsingTermDefinition
	ParsingTermExplanation
	ParsingSidenoteDefinition
	ParsingLinkDefinition
	ParsingAttributeList
	ParsingAttributeListAfterID
	ParsingAttributeListVal
	ParsingSection1
	ParsingSection1AfterAttributeList
	ParsingSection1Content
	ParsingSection2
	ParsingSection2AfterAttributeList
	ParsingSection2Content
	ParsingCodeBlock
	ParsingCodeBlockAfterAttr
	ParsingImage
	ParsingBlockquote
	ParsingBlockquoteAuthor
	ParsingBlockquoteSource
	ParsingBlockquoteAfterAttrEnd
	ParsingParagraph
	ParsingEmphasis
	ParsingStrong
	ParsingStrikethrough
	ParsingMarker
	ParsingEmphasisStrong
	ParsingEnquoteDouble
	ParsingEnquoteAngled
	ParsingLinkable
	ParsingLinkableAfterHref
	ParsingLinkableAfterRef
	ParsingSidenoteAfterRef
	ParsingSidenoteContent
)

var (
	ErrInvalidToken          = errors.New("invalid token")
	ErrSectionMissingHeading = errors.New("section must have a heading")
)

func Parse(lx LexResult) (blog *Blog, err error) {
	blog = &Blog{}
	// kinda sad how the zero value of a map isn't useable ;-(
	blog.LinkDefinitions = map[string]string{}
	blog.SidenoteDefinitions = map[string]TextRich{}
	blog.TermDefinitions = map[string]TextRich{}
	blog.Meta = Meta{}
	// parser setup
	state := ParsingStart
	levels := Levels{}
	levels.Push(&Level{ReturnToState: ParsingStart})
	var (
		currentSection1, currentSection2 *Section
		currentAttributes                = Attributes{}
		currentCodeBlock                 = &CodeBlock{Attributes: Attributes{}}
		currentImage                     = &Image{}
		currentBlockquote                = &BlockQuote{}
		currentSidenote                  = &Sidenote{}
		currentDefinition                string
		currentTerm                      string
	)
	for lexeme := range lx.Tokens() {
		level := levels.Top()
		switch state {
		default:
			panic(fmt.Errorf("parser state not implemented: %s", state))
		case ParsingStart:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenMetaBegin:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingMeta
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingDocument, Html: &Html{Name: lexeme.Text}})
				state = ParsingHtmlElement
			case lexer.TokenSection1Begin:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingSection1
			case lexer.TokenDefinitionTerm:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				currentTerm = lexeme.Text
				state = ParsingTermDefinition
			case lexer.TokenLinkDef:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				currentDefinition = lexeme.Text
				state = ParsingLinkDefinition
			case lexer.TokenSidenoteDef:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				currentDefinition = lexeme.Text
				state = ParsingSidenoteDefinition
			case lexer.TokenEOF:
				levels.Pop()
				Assert(levels.Len() == 0, "not all levels popped")
			}
		case ParsingDocument:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingDocument, Html: &Html{Name: lexeme.Text}})
				state = ParsingHtmlElement
			case lexer.TokenSection1Begin:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingSection1
			case lexer.TokenDefinitionTerm:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				currentTerm = lexeme.Text
				state = ParsingTermDefinition
			case lexer.TokenLinkDef:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				currentDefinition = lexeme.Text
				state = ParsingLinkDefinition
			case lexer.TokenSidenoteDef:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				currentDefinition = lexeme.Text
				state = ParsingSidenoteDefinition
			case lexer.TokenEOF:
				for _, c := range level.Content {
					// @todo: but about non html things? even possible?
					if h, ok := c.(*Html); ok {
						blog.Htmls = append(blog.Htmls, h)
					}
				}
				levels.Pop()
				Assert(levels.Len() == 0, "not all levels popped")
			}
		case ParsingMeta:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenMetaKey:
				level.PushString(lexeme.Text)
				state = ParsingMetaVal
			case lexer.TokenMetaEnd:
				levels.Pop()
				state = level.ReturnToState
			}
		case ParsingMetaVal:
			switch lexeme.Type {
			default:
				if !(isTextNode(lexeme) && level.TextSimple.Append(newTextNode(lexeme))) {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenMetaKey:
				// finish current key
				key := level.PopString()
				blog.Meta[key] = append(blog.Meta[key], level.TextSimple)
				level.Clear()
				// start next key
				level.PushString(lexeme.Text)
			case lexer.TokenMetaEnd:
				// finish last key
				key := level.PopString()
				blog.Meta[key] = append(blog.Meta[key], level.TextSimple)
				level.Clear()
				// finish level
				levels.Pop()
				state = level.ReturnToState
			}
		case ParsingAttributeList:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenAttributeListID:
				currentAttributes["id"] = lexeme.Text
				state = ParsingAttributeListAfterID
			case lexer.TokenAttributeListKey:
				level.PushString(lexeme.Text)
				state = ParsingAttributeListVal
			case lexer.TokenAttributeListEnd:
				levels.Pop()
				state = level.ReturnToState
			}
		case ParsingAttributeListAfterID:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenAttributeListKey:
				level.PushString(lexeme.Text)
				state = ParsingAttributeListVal
			case lexer.TokenAttributeListEnd:
				levels.Pop()
				state = level.ReturnToState
			}
		case ParsingAttributeListVal:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenText:
				level.PushString(lexeme.Text)
			case lexer.TokenAttributeListKey:
				// finish previous key
				Assert(len(level.Strings) == 1 || len(level.Strings) == 2, "")
				var val string
				if len(level.Strings) > 1 {
					val = level.PopString()
				}
				key := level.PopString()
				currentAttributes[key] = val
				level.Clear()
				// start next key
				level.PushString(lexeme.Text)
			case lexer.TokenAttributeListEnd:
				// finish last key
				Assert(len(level.Strings) == 1 || len(level.Strings) == 2, "")
				var val string
				if len(level.Strings) > 1 {
					val = level.PopString()
				}
				key := level.PopString()
				currentAttributes[key] = val
				level.Clear()
				// finish level
				levels.Pop()
				state = level.ReturnToState
			}
		case ParsingSection1:
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case isTextNode(lexeme):
				ok := level.TextRich.Append(newTextNode(lexeme))
				Assert(ok, "all text nodes must fit into rich text")
			case lexeme.Type == lexer.TokenAttributeListBegin:
				levels.Push(&Level{ReturnToState: ParsingSection1AfterAttributeList})
				state = ParsingAttributeList
			case lexeme.Type == lexer.TokenSection1Content:
				if len(level.TextRich) == 0 {
					err = errors.Join(err, newError(lexeme, state, ErrSectionMissingHeading))
				}
				currentSection1 = &Section{
					Level:   1,
					Heading: level.TextRich,
				}
				level.Clear()
				state = ParsingSection1Content
			}
		case ParsingSection1AfterAttributeList:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenSection1Content:
				if len(level.TextRich) == 0 {
					err = errors.Join(err, newError(lexeme, state, ErrSectionMissingHeading))
				}
				currentSection1 = &Section{
					Attributes: currentAttributes,
					Level:      1,
					Heading:    level.TextRich,
				}
				level.Clear()
				currentAttributes = Attributes{}
				state = ParsingSection1Content
			}
		case ParsingSection1Content:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenDefinitionTerm:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				currentTerm = lexeme.Text
				state = ParsingTermDefinition
			case lexer.TokenHorizontalRule:
				level.Content = append(level.Content, &HorizontalRule{})
			case lexer.TokenCodeBlockBegin:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingCodeBlock
			case lexer.TokenSection2Begin:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingSection2
			case lexer.TokenImageBegin:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingImage
			case lexer.TokenBlockquoteBegin:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				currentBlockquote = &BlockQuote{}
				state = ParsingBlockquote
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingSection1Content, Html: &Html{Name: lexeme.Text}})
				state = ParsingHtmlElement
			case lexer.TokenLinkDef:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				currentDefinition = lexeme.Text
				state = ParsingLinkDefinition
			case lexer.TokenSidenoteDef:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				currentDefinition = lexeme.Text
				state = ParsingSidenoteDefinition
			case lexer.TokenParagraphBegin:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingParagraph
			case lexer.TokenSection1End:
				currentSection1.Content = level.Content
				blog.Sections = append(blog.Sections, currentSection1)
				currentSection1 = &Section{}
				levels.Pop()
				state = level.ReturnToState
			}
		case ParsingSection2:
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case isTextNode(lexeme):
				ok := level.TextRich.Append(newTextNode(lexeme))
				Assert(ok, "all text nodes must fit into rich text")
			case lexeme.Type == lexer.TokenAttributeListBegin:
				levels.Push(&Level{ReturnToState: ParsingSection2AfterAttributeList})
				state = ParsingAttributeList
			case lexeme.Type == lexer.TokenSection2Content:
				if len(level.TextRich) == 0 {
					err = errors.Join(err, newError(lexeme, state, ErrSectionMissingHeading))
				}
				currentSection2 = &Section{
					Level:   2,
					Heading: level.TextRich,
				}
				level.Clear()
				state = ParsingSection2Content
			}
		case ParsingSection2AfterAttributeList:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenSection2Content:
				if len(level.TextRich) == 0 {
					err = errors.Join(err, newError(lexeme, state, ErrSectionMissingHeading))
				}
				currentSection2 = &Section{
					Attributes: currentAttributes,
					Level:      2,
					Heading:    level.TextRich,
				}
				level.Clear()
				currentAttributes = Attributes{}
				state = ParsingSection2Content
			}
		case ParsingSection2Content:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenDefinitionTerm:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				currentTerm = lexeme.Text
				state = ParsingTermDefinition
			case lexer.TokenHorizontalRule:
				level.Content = append(level.Content, &HorizontalRule{})
			case lexer.TokenCodeBlockBegin:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				state = ParsingCodeBlock
			case lexer.TokenImageBegin:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				state = ParsingImage
			case lexer.TokenBlockquoteBegin:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				currentBlockquote = &BlockQuote{}
				state = ParsingBlockquote
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingSection2Content, Html: &Html{Name: lexeme.Text}})
				state = ParsingHtmlElement
			case lexer.TokenLinkDef:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				currentDefinition = lexeme.Text
				state = ParsingLinkDefinition
			case lexer.TokenSidenoteDef:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				currentDefinition = lexeme.Text
				state = ParsingSidenoteDefinition
			case lexer.TokenParagraphBegin:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				state = ParsingParagraph
			case lexer.TokenSection2End:
				currentSection2.Content = level.Content
				levels.Pop()
				parent := levels.Top()
				parent.Content = append(parent.Content, currentSection2)
				currentSection2 = &Section{}
				state = level.ReturnToState
			}
		case ParsingParagraph:
			// @todo: case lexer.TokenEnquoteSingleBegin:
			switch lexeme.Type {
			default:
				if !(isTextNode(lexeme) && level.TextRich.Append(newTextNode(lexeme))) {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenEmphasisBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingEmphasis
			case lexer.TokenStrongBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingStrong
			case lexer.TokenEmphasisStrongBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingEmphasisStrong
			case lexer.TokenEnquoteDoubleBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingEnquoteDouble
			case lexer.TokenEnquoteAngledBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingEnquoteAngled
			case lexer.TokenStrikethroughBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingStrikethrough
			case lexer.TokenMarkerBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingMarker
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingParagraph, Html: &Html{Name: lexeme.Text}})
				state = ParsingHtmlElement
			case lexer.TokenLinkableBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingLinkable
			case lexer.TokenParagraphEnd:
				levels.Pop()
				parent := levels.Top()
				parent.Content = append(parent.Content, &Paragraph{
					Content: level.TextRich,
				})
				state = level.ReturnToState
			}
		case ParsingEnquoteDouble:
			switch lexeme.Type {
			default:
				if !(isTextNode(lexeme) && level.TextRich.Append(newTextNode(lexeme))) {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenEmphasisBegin:
				levels.Push(&Level{ReturnToState: ParsingEnquoteDouble})
				state = ParsingEmphasis
			case lexer.TokenEmphasisStrongBegin:
				levels.Push(&Level{ReturnToState: ParsingEnquoteDouble})
				state = ParsingEmphasisStrong
			case lexer.TokenStrongBegin:
				levels.Push(&Level{ReturnToState: ParsingEnquoteDouble})
				state = ParsingStrong
			case lexer.TokenStrikethroughBegin:
				levels.Push(&Level{ReturnToState: ParsingEnquoteDouble})
				state = ParsingStrikethrough
			case lexer.TokenMarkerBegin:
				levels.Push(&Level{ReturnToState: ParsingEnquoteDouble})
				state = ParsingMarker
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingEnquoteDouble})
				state = ParsingHtmlElement
			case lexer.TokenLinkableBegin:
				levels.Push(&Level{ReturnToState: ParsingEnquoteDouble})
				state = ParsingLinkable
				// @todo: EnquoteAngled?
			case lexer.TokenEnquoteDoubleEnd:
				levels.Pop()
				parent := levels.Top()
				ok := parent.TextRich.Append(AsRef(EnquoteDouble(level.TextRich)))
				Assert(ok, "enquote double must be accepted as rich text")
				state = level.ReturnToState
			}
		case ParsingEnquoteAngled:
			switch lexeme.Type {
			default:
				if !(isTextNode(lexeme) && level.TextRich.Append(newTextNode(lexeme))) {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenEmphasisBegin:
				levels.Push(&Level{ReturnToState: ParsingEnquoteAngled})
				state = ParsingEmphasis
			case lexer.TokenEmphasisStrongBegin:
				levels.Push(&Level{ReturnToState: ParsingEnquoteAngled})
				state = ParsingEmphasisStrong
			case lexer.TokenStrongBegin:
				levels.Push(&Level{ReturnToState: ParsingEnquoteAngled})
				state = ParsingStrong
			case lexer.TokenStrikethroughBegin:
				levels.Push(&Level{ReturnToState: ParsingEnquoteAngled})
				state = ParsingStrikethrough
			case lexer.TokenMarkerBegin:
				levels.Push(&Level{ReturnToState: ParsingEnquoteAngled})
				state = ParsingMarker
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingEnquoteAngled})
				state = ParsingHtmlElement
			case lexer.TokenLinkableBegin:
				levels.Push(&Level{ReturnToState: ParsingEnquoteAngled})
				state = ParsingLinkable
			case lexer.TokenEnquoteAngledEnd:
				levels.Pop()
				parent := levels.Top()
				ok := parent.TextRich.Append(AsRef(EnquoteAngled(level.TextRich)))
				Assert(ok, "enquote angled must be accepted as rich text")
				state = level.ReturnToState
			}
		case ParsingEmphasis:
			switch lexeme.Type {
			default:
				if !(isTextNode(lexeme) && level.TextRich.Append(newTextNode(lexeme))) {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenEmphasisStrongBegin:
				levels.Push(&Level{ReturnToState: ParsingEmphasis})
				state = ParsingStrong // instead of ParsingEmphasisStrong, because we're already inside an emphasis
			case lexer.TokenStrongBegin:
				levels.Push(&Level{ReturnToState: ParsingEmphasis})
				state = ParsingStrong
			case lexer.TokenStrikethroughBegin:
				levels.Push(&Level{ReturnToState: ParsingEmphasis})
				state = ParsingStrikethrough
			case lexer.TokenMarkerBegin:
				levels.Push(&Level{ReturnToState: ParsingEmphasis})
				state = ParsingMarker
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingEmphasis})
				state = ParsingHtmlElement
			case lexer.TokenLinkableBegin:
				levels.Push(&Level{ReturnToState: ParsingEmphasis})
				state = ParsingLinkable
			case lexer.TokenEnquoteDoubleBegin:
				levels.Push(&Level{ReturnToState: ParsingEmphasis})
				state = ParsingEnquoteDouble
			case lexer.TokenEnquoteAngledBegin:
				levels.Push(&Level{ReturnToState: ParsingEmphasis})
				state = ParsingEnquoteAngled
			case lexer.TokenEmphasisEnd:
				levels.Pop()
				parent := levels.Top()
				ok := parent.TextRich.Append(AsRef(Emphasis(level.TextRich)))
				Assert(ok, "emphasis must be accepted as rich text")
				state = level.ReturnToState
			}
		case ParsingStrong:
			switch lexeme.Type {
			default:
				if !(isTextNode(lexeme) && level.TextRich.Append(newTextNode(lexeme))) {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenEmphasisStrongBegin:
				levels.Push(&Level{ReturnToState: ParsingStrong})
				state = ParsingEmphasis // intead of ParsingEmphasisStrong, because we're already inside a strong
			case lexer.TokenEmphasisBegin:
				levels.Push(&Level{ReturnToState: ParsingStrong})
				state = ParsingEmphasis
			case lexer.TokenStrikethroughBegin:
				levels.Push(&Level{ReturnToState: ParsingStrong})
				state = ParsingStrikethrough
			case lexer.TokenMarkerBegin:
				levels.Push(&Level{ReturnToState: ParsingStrong})
				state = ParsingMarker
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingStrong})
				state = ParsingHtmlElement
			case lexer.TokenLinkableBegin:
				levels.Push(&Level{ReturnToState: ParsingStrong})
				state = ParsingLinkable
			case lexer.TokenEnquoteDoubleBegin:
				levels.Push(&Level{ReturnToState: ParsingStrong})
				state = ParsingEnquoteDouble
			case lexer.TokenEnquoteAngledBegin:
				levels.Push(&Level{ReturnToState: ParsingStrong})
				state = ParsingEnquoteAngled
			case lexer.TokenStrongEnd:
				levels.Pop()
				parent := levels.Top()
				ok := parent.TextRich.Append(AsRef(Strong(level.TextRich)))
				Assert(ok, "strong must be accepted as rich text")
				state = level.ReturnToState
			}
		case ParsingEmphasisStrong:
			switch lexeme.Type {
			default:
				if !(isTextNode(lexeme) && level.TextRich.Append(newTextNode(lexeme))) {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenStrongBegin:
				// ignore the token
			case lexer.TokenStrongEnd:
				// ignore the token
			case lexer.TokenEmphasisBegin:
				// ignore the token
			case lexer.TokenEmphasisEnd:
				// ignore the token
			case lexer.TokenStrikethroughBegin:
				levels.Push(&Level{ReturnToState: ParsingEmphasisStrong})
				state = ParsingStrikethrough
			case lexer.TokenMarkerBegin:
				levels.Push(&Level{ReturnToState: ParsingEmphasisStrong})
				state = ParsingMarker
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingEmphasisStrong})
				state = ParsingHtmlElement
			case lexer.TokenLinkableBegin:
				levels.Push(&Level{ReturnToState: ParsingEmphasisStrong})
				state = ParsingLinkable
			case lexer.TokenEnquoteDoubleBegin:
				levels.Push(&Level{ReturnToState: ParsingEmphasisStrong})
				state = ParsingEnquoteDouble
			case lexer.TokenEnquoteAngledBegin:
				levels.Push(&Level{ReturnToState: ParsingEmphasisStrong})
				state = ParsingEnquoteAngled
			case lexer.TokenEmphasisStrongEnd:
				levels.Pop()
				parent := levels.Top()
				ok := parent.TextRich.Append(AsRef(EmphasisStrong(level.TextRich)))
				Assert(ok, "emphasis strong must be accepted as rich text")
				state = level.ReturnToState
			}
		case ParsingStrikethrough:
			switch lexeme.Type {
			default:
				if !(isTextNode(lexeme) && level.TextRich.Append(newTextNode(lexeme))) {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenEmphasisBegin:
				levels.Push(&Level{ReturnToState: ParsingStrikethrough})
				state = ParsingEmphasis
			case lexer.TokenStrongBegin:
				levels.Push(&Level{ReturnToState: ParsingStrikethrough})
				state = ParsingStrong
			case lexer.TokenEmphasisStrongBegin:
				levels.Push(&Level{ReturnToState: ParsingStrikethrough})
				state = ParsingEmphasisStrong
			case lexer.TokenEnquoteDoubleBegin:
				levels.Push(&Level{ReturnToState: ParsingStrikethrough})
				state = ParsingEnquoteDouble
			case lexer.TokenEnquoteAngledBegin:
				levels.Push(&Level{ReturnToState: ParsingStrikethrough})
				state = ParsingEnquoteAngled
			case lexer.TokenMarkerBegin:
				levels.Push(&Level{ReturnToState: ParsingStrikethrough})
				state = ParsingMarker
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingStrikethrough, Html: &Html{Name: lexeme.Text}})
				state = ParsingHtmlElement
			case lexer.TokenLinkableBegin:
				levels.Push(&Level{ReturnToState: ParsingStrikethrough})
				state = ParsingLinkable
			case lexer.TokenStrikethroughEnd:
				levels.Pop()
				parent := levels.Top()
				ok := parent.TextRich.Append(AsRef(Strikethrough(level.TextRich)))
				Assert(ok, "strikethrough must be accepted as rich text")
				state = level.ReturnToState
			}
		case ParsingMarker:
			switch lexeme.Type {
			default:
				if !(isTextNode(lexeme) && level.TextRich.Append(newTextNode(lexeme))) {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenEmphasisBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingEmphasis
			case lexer.TokenStrongBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingStrong
			case lexer.TokenEmphasisStrongBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingEmphasisStrong
			case lexer.TokenEnquoteDoubleBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingEnquoteDouble
			case lexer.TokenEnquoteAngledBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingEnquoteAngled
			case lexer.TokenStrikethroughBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingStrikethrough
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingParagraph, Html: &Html{Name: lexeme.Text}})
				state = ParsingHtmlElement
			case lexer.TokenLinkableBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingLinkable
			case lexer.TokenMarkerEnd:
				levels.Pop()
				parent := levels.Top()
				ok := parent.TextRich.Append(AsRef(Marker(level.TextRich)))
				Assert(ok, "marker must be accepted as rich text")
				state = level.ReturnToState
			}
		case ParsingLinkable:
			switch lexeme.Type {
			default:
				if !(isTextNode(lexeme) && level.TextRich.Append(newTextNode(lexeme))) {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenLinkHref:
				level.PushString(lexeme.Text)
				state = ParsingLinkableAfterHref
			case lexer.TokenLinkRef:
				level.PushString(lexeme.Text)
				state = ParsingLinkableAfterRef
			case lexer.TokenSidenoteRef:
				currentSidenote.Word = level.TextRich
				currentSidenote.Ref = lexeme.Text
				//level.Clear() @todo: ?
				state = ParsingSidenoteAfterRef
			case lexer.TokenSidenoteContent:
				currentSidenote.Word = level.TextRich
				level.Clear()
				state = ParsingSidenoteContent
			case lexer.TokenLinkableEnd:
				levels.Pop()
				parent := levels.Top()
				ok := parent.TextRich.Append(AsRef(Link{
					Name: level.TextRich,
				}))
				Assert(ok, "link must be accepted as rich text")
				state = level.ReturnToState
			}
		case ParsingLinkableAfterHref:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenLinkableEnd:
				levels.Pop()
				parent := levels.Top()
				ok := parent.TextRich.Append(AsRef(Link{
					Href: level.PopString(),
					Name: level.TextRich,
				}))
				Assert(ok, "link must be accepted as rich text")
				state = level.ReturnToState
			}
		case ParsingLinkableAfterRef:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenLinkableEnd:
				levels.Pop()
				parent := levels.Top()
				ok := parent.TextRich.Append(AsRef(Link{
					Ref:  level.PopString(),
					Name: level.TextRich,
				}))
				Assert(ok, "link must be accepted as rich text")
				state = level.ReturnToState
			}
		case ParsingSidenoteAfterRef:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenLinkableEnd:
				levels.Pop()
				parent := levels.Top()
				ok := parent.TextRich.Append(currentSidenote)
				Assert(ok, "sidenote must be accepted as rich text")
				currentSidenote = &Sidenote{}
				state = level.ReturnToState
			}
		case ParsingSidenoteContent:
			switch lexeme.Type {
			default:
				if !(isTextNode(lexeme) && level.TextRich.Append(newTextNode(lexeme))) {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenLinkableEnd:
				levels.Pop()
				parent := levels.Top()
				currentSidenote.Content = level.TextRich
				ok := parent.TextRich.Append(currentSidenote)
				Assert(ok, "sidenote must be accepted as rich text")
				currentSidenote = &Sidenote{}
				state = level.ReturnToState
			}
		case ParsingCodeBlock:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenCodeBlockLang:
				currentCodeBlock.Attributes["Lang"] = lexeme.Text
			case lexer.TokenAttributeListBegin:
				levels.Push(&Level{ReturnToState: ParsingCodeBlockAfterAttr})
				state = ParsingAttributeList
			case lexer.TokenText:
				level.PushString(lexeme.Text)
				state = ParsingCodeBlockAfterAttr
			case lexer.TokenCodeBlockEnd:
				// empty code block
				levels.Pop()
				parent := levels.Top()
				parent.Content = append(parent.Content, currentCodeBlock)
				currentCodeBlock = &CodeBlock{Attributes: Attributes{}}
				state = level.ReturnToState
			}
		case ParsingCodeBlockAfterAttr:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenText:
				level.PushString(lexeme.Text)
			case lexer.TokenCodeBlockEnd:
				levels.Pop()
				parent := levels.Top()
				currentCodeBlock.Lines = level.Strings
				parent.Content = append(parent.Content, currentCodeBlock)
				currentCodeBlock = &CodeBlock{Attributes: Attributes{}}
				state = level.ReturnToState
			}
		case ParsingImage:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenImageAltText:
				currentImage.Alt = TextSimple{newTextNode(lexeme)}
			case lexer.TokenImagePath:
				currentImage.Name = lexeme.Text
			case lexer.TokenImageTitle:
				currentImage.Title = TextSimple{newTextNode(lexeme)}
			case lexer.TokenImageEnd:
				levels.Pop()
				parent := levels.Top()
				parent.Content = append(parent.Content, currentImage)
				currentImage = &Image{}
				state = level.ReturnToState
			}
		case ParsingBlockquote:
			switch lexeme.Type {
			default:
				if !(isTextNode(lexeme) && level.TextRich.Append(newTextNode(lexeme))) {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenEmphasisBegin:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingEmphasis
			case lexer.TokenStrongBegin:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingStrong
			case lexer.TokenEmphasisStrongBegin:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingEmphasisStrong
			case lexer.TokenEnquoteDoubleBegin:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingEnquoteDouble
			case lexer.TokenEnquoteAngledBegin:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingEnquoteAngled
			case lexer.TokenStrikethroughBegin:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingStrikethrough
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingBlockquote, Html: &Html{Name: lexeme.Text}})
				state = ParsingHtmlElement
			case lexer.TokenLinkableBegin:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingLinkable
			case lexer.TokenMarkerBegin:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingMarker
			case lexer.TokenBlockquoteAttrAuthor:
				currentBlockquote.QuoteText = level.TextRich
				level.Clear()
				state = ParsingBlockquoteAuthor
			case lexer.TokenBlockquoteEnd:
				levels.Pop()
				parent := levels.Top()
				currentBlockquote.QuoteText = level.TextRich
				parent.Content = append(parent.Content, currentBlockquote)
				state = level.ReturnToState
			}
		case ParsingBlockquoteAuthor:
			switch lexeme.Type {
			default:
				if !(isTextNode(lexeme) && level.TextSimple.Append(newTextNode(lexeme))) {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenBlockquoteAttrSource:
				currentBlockquote.Author = level.TextSimple
				level.Clear()
				state = ParsingBlockquoteSource
			case lexer.TokenBlockquoteAttrEnd:
				currentBlockquote.Author = level.TextSimple
				level.Clear()
				state = ParsingBlockquoteAfterAttrEnd
			}
		case ParsingBlockquoteSource:
			switch lexeme.Type {
			default:
				if !(isTextNode(lexeme) && level.TextRich.Append(newTextNode(lexeme))) {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenLinkableBegin:
				levels.Push(&Level{ReturnToState: ParsingBlockquoteSource})
				state = ParsingLinkable
			case lexer.TokenBlockquoteAttrEnd:
				currentBlockquote.Source = level.TextRich
				level.Clear()
				state = ParsingBlockquoteAfterAttrEnd
			}
		case ParsingBlockquoteAfterAttrEnd:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenBlockquoteEnd:
				levels.Pop()
				parent := levels.Top()
				parent.Content = append(parent.Content, currentBlockquote)
				state = level.ReturnToState
			}
		case ParsingLinkDefinition:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenText:
				blog.LinkDefinitions[currentDefinition] = lexeme.Text
				levels.Pop()
				state = level.ReturnToState
			}
		case ParsingSidenoteDefinition:
			switch lexeme.Type {
			default:
				if !(isTextNode(lexeme) && level.TextRich.Append(newTextNode(lexeme))) {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenSidenoteDefEnd:
				blog.SidenoteDefinitions[currentDefinition] = level.TextRich
				levels.Pop()
				state = level.ReturnToState
			}
		case ParsingTermDefinition:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenDefinitionExplanationBegin:
				state = ParsingTermExplanation
			}
		case ParsingTermExplanation:
			switch lexeme.Type {
			default:
				if !(isTextNode(lexeme) && level.TextRich.Append(newTextNode(lexeme))) {
					err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
				}
			case lexer.TokenDefinitionExplanationEnd:
				levels.Pop()
				blog.TermDefinitions[currentTerm] = level.TextRich
				currentTerm = ""
				state = level.ReturnToState
			}
		case ParsingHtmlElement:
			switch lexeme.Type {
			default:
			case lexer.TokenHtmlTagAttrKey:
				level.PushString(lexeme.Text)
				level.Html.Attributes = Attributes{}
				state = ParsingHtmlElementAttributes
			case lexer.TokenHtmlTagContent:
				state = ParsingHtmlElementContent
			}
		case ParsingHtmlElementAttributes:
			switch lexeme.Type {
			default:
			case lexer.TokenHtmlTagAttrKey:
				// finish current key
				var val string
				if len(level.Strings) > 0 {
					Assert(len(level.Strings) == 2, "")
					val = level.PopString()
				}
				key := level.PopString()
				level.Html.Attributes[key] = val
				// start next key
				level.PushString(lexeme.Text)
			case lexer.TokenHtmlTagAttrVal:
				level.PushString(lexeme.Text)
			case lexer.TokenHtmlTagContent:
				// finish last key
				var val string
				if len(level.Strings) > 0 {
					Assert(len(level.Strings) == 2, "")
					val = level.PopString()
				}
				key := level.PopString()
				level.Html.Attributes[key] = val
				// go to parsing element content
				state = ParsingHtmlElementContent
			}
		case ParsingHtmlElementContent:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, ErrInvalidToken))
			case lexer.TokenParagraphBegin:
				levels.Push(&Level{ReturnToState: ParsingHtmlElementContent})
				state = ParsingParagraph
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingHtmlElementContent, Html: &Html{Name: lexeme.Text}})
				state = ParsingHtmlElement
			case lexer.TokenHtmlTagClose:
				levels.Pop()
				parent := levels.Top()
				level.Html.Content = level.Content
				parent.Content = append(parent.Content, level.Html)
				state = level.ReturnToState
			}
		}
	}
	return
}
