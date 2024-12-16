package parser

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"strconv"
	"log"

	"github.com/kr/pretty"

	. "github.com/cvanloo/blog-go/assert"
	"github.com/cvanloo/blog-go/markup/lexer"
	"github.com/cvanloo/blog-go/markup/gen"
)

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

func newError(lexeme lexer.Token, state ParseState, inner error) ParserError {
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
		Strings Stack[string]
		Content Stack[gen.Renderable]
		TextValues Stack[gen.StringRenderable]
	}
	Levels struct {
		levels []*Level
	}
)

func (lv *Level) PushString(v string) {
	lv.Strings = lv.Strings.Push(v)
}

func (lv *Level) PopString() (s string) {
	lv.Strings, s = lv.Strings.Pop()
	return s
}

func (lv *Level) PushContent(v gen.Renderable) {
	lv.Content = lv.Content.Push(v)
}

func (lv *Level) PushText(v gen.StringRenderable) {
	lv.TextValues = lv.TextValues.Push(v)
}

func (lv *Level) EmptyText() {
	lv.TextValues = lv.TextValues.Empty()
}

func (ls *Levels) Push(l *Level) {
	ls.levels = append(ls.levels, l)
}

func (ls *Levels) Top() *Level {
	l := len(ls.levels)
	return ls.levels[l-1]
}

func (ls *Levels) Dig() *Level {
	l := len(ls.levels)
	return ls.levels[l-2]
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

//go:generate stringer -type ParseState
type ParseState int

const (
	ParsingStart ParseState = iota
	ParsingMeta
	ParsingMetaVal
	ParsingDocument
	ParsingSection1
	ParsingSection1Content
	ParsingSection2
	ParsingSection2Content
	ParsingHtmlTag
	ParsingParagraph
	ParsingCodeBlock
	ParsingImage
	ParsingImageTitle
	ParsingImageAlt
	ParsingBlockquote
	ParsingBlockquoteAttrAuthor
	ParsingBlockquoteAttrSource
	ParsingEnquote
	ParsingEmphasis
	ParsingStrong
	ParsingEmphasisStrong
	ParsingLink
)

func Parse(lx LexResult) (blog gen.Blog, err error) {
	state := ParsingStart
	levels := Levels{}
	var (
		currentBlockquote = gen.Blockquote{}
		currentImage = gen.Image{}
		//currentSection1 = gen.Section{Level: 1}
		//currentSection2 = gen.Section{Level: 2}
	)
	for lexeme := range lx.Tokens() {
		log.Printf("[%s/%s] %# v", state, lexeme, pretty.Formatter(levels))
		switch state {
		default:
			panic(fmt.Errorf("parser state not implemented: %s", state))
		case ParsingStart:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token, expected one of MetaBegin, HtmlTagOpen, Section1")))
			case lexer.TokenMetaBegin:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingMeta
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingHtmlTag
			case lexer.TokenSection1Begin:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingSection1
			}
		case ParsingMeta:
			level := levels.Top()
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token, expected one of MetaKey, MetaEnd")))
			case lexer.TokenMetaKey:
				if !checkMetaKey(lexeme.Text) {
					err = errors.Join(err, newError(lexeme, state, fmt.Errorf("unrecognized meta key: %s", lexeme.Text)))
				}
				level.PushString(lexeme.Text)
				state = ParsingMetaVal
			case lexer.TokenMetaEnd:
				_ = levels.Pop()
				state = level.ReturnToState
			}
		case ParsingMetaVal:
			level := levels.Top()
			switch lexeme.Type {
			default:
				if isTextContent(lexeme.Type) {
					level.PushText(newTextContent(lexeme))
				} else {
					err = errors.Join(err, newError(lexeme, state, errors.New("invalid token, expected one of MetaKey, MetaEnd, string content")))
				}
			case lexer.TokenMetaKey:
				metaKey := level.PopString()
				setMetaKeyValuePair(&blog, metaKey, gen.StringOnlyContent(level.TextValues))
				level.EmptyText()
				if !checkMetaKey(lexeme.Text) {
					err = errors.Join(err, newError(lexeme, state, fmt.Errorf("unrecognized meta key: %s", lexeme.Text)))
				}
				level.PushString(lexeme.Text)
			case lexer.TokenMetaEnd:
				metaKey := level.PopString()
				setMetaKeyValuePair(&blog, metaKey, gen.StringOnlyContent(level.TextValues))
				level.EmptyText()
				_ = levels.Pop()
				state = level.ReturnToState
			}
		case ParsingDocument:
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token, expected one of Section1, HtmlTagOpen")))
			case lexer.TokenSection1Begin:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingSection1
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingDocument})
				state = ParsingHtmlTag
			case lexer.TokenEOF:
				// @todo
			}
		case ParsingSection1:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token, expected one of Section1Content or text content")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenSection1Content:
				if len(level.TextValues) == 0 {
					err = errors.Join(err, newError(lexeme, state, errors.New("section must have a heading")))
				}
				blog.Sections = append(blog.Sections, gen.Section{
					Level: 1,
					Heading: gen.StringOnlyContent(level.TextValues),
				})
				level.EmptyText()
				//levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingSection1Content
			}
		case ParsingSection1Content:
			level := levels.Top()
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case lexer.TokenSection1End:
				_ = levels.Pop()
				l := len(blog.Sections)
				blog.Sections[l-1].Content = level.Content
				state = level.ReturnToState
			case lexer.TokenSection2Begin:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingSection2
			case lexer.TokenParagraphBegin:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingParagraph
			case lexer.TokenCodeBlockBegin:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingCodeBlock
			case lexer.TokenImageBegin:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingImage
			case lexer.TokenBlockquoteBegin:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingBlockquote
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingSection1Content})
				state = ParsingHtmlTag
			case lexer.TokenHorizontalRule:
				level.PushContent(gen.HorizontalRule{})
			}
		case ParsingSection2:
			level := levels.Top()
			Assert(level.ReturnToState == ParsingSection1Content, "section 2 must appear within a section 1")
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token, expected one of Section2Content or text content")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenSection2Content:
				if len(level.TextValues) == 0 {
					err = errors.Join(err, newError(lexeme, state, errors.New("section must have a heading")))
				}
				paren := levels.Dig()
				paren.PushContent(gen.Section{
					Level: 2,
					Heading: gen.StringOnlyContent(level.TextValues),
				})
				level.EmptyText()
				//levels.Push(&Level{ReturnToState: ParsingSection2Content})
				state = ParsingSection2Content
			}
		case ParsingSection2Content:
			level := levels.Top()
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case lexer.TokenSection2End:
				_ = levels.Pop()
				paren := levels.Top()
				l := len(paren.Content)
				section := paren.Content[l-1].(gen.Section) // @todo: I really don't like this
				section.Content = level.Content
				paren.Content[l-1] = section
				Assert(level.ReturnToState == ParsingSection1Content, "confused parser state")
				state = level.ReturnToState
			case lexer.TokenParagraphBegin:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				state = ParsingParagraph
			case lexer.TokenCodeBlockBegin:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				state = ParsingCodeBlock
			case lexer.TokenImageBegin:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				state = ParsingImage
			case lexer.TokenBlockquoteBegin:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				state = ParsingBlockquote
			case lexer.TokenHtmlTagOpen:
				levels.Push(&Level{ReturnToState: ParsingSection2Content})
				state = ParsingHtmlTag
			case lexer.TokenHorizontalRule:
				level.PushContent(gen.HorizontalRule{})
			}
		case ParsingParagraph:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token, expected one of ParagraphEnd, or text content")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenParagraphEnd:
				_ = levels.Pop()
				paren := levels.Top()
				paren.PushContent(gen.Paragraph{
					Content: gen.StringOnlyContent(level.TextValues),
				})
				state = level.ReturnToState
			case lexeme.Type == lexer.TokenEmphasis:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingEmphasis
			case lexeme.Type == lexer.TokenStrong:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingStrong
			case lexeme.Type == lexer.TokenEmphasisStrong:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingEmphasisStrong
			case lexeme.Type == lexer.TokenEnquoteBegin:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingEnquote
			case lexeme.Type == lexer.TokenLink:
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingLink
			case lexeme.Type == lexer.TokenHtmlTagOpen:
				// @todo: this is a tad bit tricky (StringRenderable is part of the paragraph, Renderable ends the paragraph, and comes after it)
				levels.Push(&Level{ReturnToState: ParsingParagraph})
				state = ParsingHtmlTag
			}
		case ParsingHtmlTag:
			switch lexeme.Type {
			default:
				// @todo
			case lexer.TokenHtmlTagClose:
				level := levels.Pop()
				state = level.ReturnToState
			}
		case ParsingCodeBlock:
			level := levels.Top()
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case lexer.TokenCodeBlockLang: // @todo
			case lexer.TokenCodeBlockLineFirst:
			case lexer.TokenCodeBlockLineLast:
			case lexer.TokenText:
				level.PushString(lexeme.Text)
			case lexer.TokenCodeBlockEnd:
				_ = levels.Pop()
				paren := levels.Top()
				paren.PushContent(gen.CodeBlock{
					Lines: level.Strings,
				})
				state = level.ReturnToState
			}
		case ParsingImage:
			level := levels.Top()
			switch lexeme.Type {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case lexer.TokenImageTitle:
				levels.Push(&Level{ReturnToState: ParsingImage})
				state = ParsingImageTitle
			case lexer.TokenImageAlt:
				levels.Push(&Level{ReturnToState: ParsingImage})
				state = ParsingImageAlt
			case lexer.TokenImagePath:
				currentImage.Name = lexeme.Text
			case lexer.TokenImageEnd:
				_ = levels.Pop()
				paren := levels.Top()
				paren.PushContent(currentImage)
				currentImage = gen.Image{}
				state = level.ReturnToState
			}
		case ParsingImageTitle:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenImageAttrEnd:
				_ = levels.Pop()
				currentImage.Title = gen.StringOnlyContent(level.TextValues)
				Assert(level.ReturnToState == ParsingImage, "confused parser state")
				state = level.ReturnToState
			}
		case ParsingImageAlt:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenImageAttrEnd:
				_ = levels.Pop()
				currentImage.Alt = gen.StringOnlyContent(level.TextValues)
				Assert(level.ReturnToState == ParsingImage, "confused parser state")
				state = level.ReturnToState
			}
		case ParsingBlockquote:
			level := levels.Top()
			// @todo: also allow links inside blockquote
			switch lexeme.Type {
			default:
				if isTextContent(lexeme.Type) {
					level.PushText(newTextContent(lexeme))
				} else {
					err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
				}
			case lexer.TokenBlockquoteAttrAuthor:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingBlockquoteAttrAuthor
			case lexer.TokenBlockquoteAttrSource:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingBlockquoteAttrSource
			case lexer.TokenEmphasis:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingEmphasis
			case lexer.TokenStrong:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingStrong
			case lexer.TokenEmphasisStrong:
				levels.Push(&Level{ReturnToState: ParsingBlockquote})
				state = ParsingEmphasisStrong
			case lexer.TokenBlockquoteEnd:
				currentBlockquote.QuoteText = gen.StringOnlyContent(level.TextValues)
				_ = levels.Pop()
				paren := levels.Top()
				paren.PushContent(currentBlockquote)
				currentBlockquote = gen.Blockquote{}
				state = level.ReturnToState
			}
		case ParsingBlockquoteAttrAuthor:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenBlockquoteAttrEnd:
				_ = levels.Pop()
				currentBlockquote.Author = gen.StringOnlyContent(level.TextValues)
				Assert(level.ReturnToState == ParsingBlockquote, "confused parser state")
				state = level.ReturnToState
			}
		case ParsingBlockquoteAttrSource:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenBlockquoteAttrEnd:
				_ = levels.Pop()
				currentBlockquote.Source = gen.StringOnlyContent(level.TextValues)
				Assert(level.ReturnToState == ParsingBlockquote, "confused parser state")
				state = level.ReturnToState
			}
		case ParsingEnquote:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("enquote can only contain text content")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenEmphasis:
				levels.Push(&Level{ReturnToState: ParsingEnquote})
				state = ParsingEmphasis
			case lexeme.Type == lexer.TokenStrong:
				levels.Push(&Level{ReturnToState: ParsingEnquote})
				state = ParsingStrong
			case lexeme.Type == lexer.TokenEmphasisStrong:
				levels.Push(&Level{ReturnToState: ParsingEnquote})
				state = ParsingEmphasisStrong
			case lexeme.Type == lexer.TokenEnquoteEnd:
				level := levels.Pop()
				paren := levels.Top()
				paren.PushText(gen.Enquote{
					gen.StringOnlyContent(level.TextValues),
				})
				state = level.ReturnToState
			}
		case ParsingEmphasis:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenStrong:
				levels.Push(&Level{ReturnToState: ParsingEmphasis})
				state = ParsingStrong
			case lexeme.Type == lexer.TokenEmphasisStrong:
				levels.Push(&Level{ReturnToState: ParsingEmphasis})
				state = ParsingEmphasisStrong
			case lexeme.Type == lexer.TokenEmphasis:
				_ = levels.Pop()
				paren := levels.Top()
				paren.PushText(gen.Emphasis{
					gen.StringOnlyContent(level.TextValues),
				})
				state = level.ReturnToState
			}
		case ParsingStrong:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenEmphasis:
				levels.Push(&Level{ReturnToState: ParsingStrong})
				state = ParsingStrong
			case lexeme.Type == lexer.TokenEmphasisStrong:
				levels.Push(&Level{ReturnToState: ParsingStrong})
				state = ParsingEmphasisStrong
			case lexeme.Type == lexer.TokenStrong:
				_ = levels.Pop()
				paren := levels.Top()
				paren.PushText(gen.Strong{
					gen.StringOnlyContent(level.TextValues),
				})
				state = level.ReturnToState
			}
		case ParsingEmphasisStrong:
			level := levels.Top()
			switch {
			default:
				err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
			case isTextContent(lexeme.Type):
				level.PushText(newTextContent(lexeme))
			case lexeme.Type == lexer.TokenEmphasis:
				levels.Push(&Level{ReturnToState: ParsingEmphasisStrong})
				state = ParsingStrong
			case lexeme.Type == lexer.TokenStrong:
				levels.Push(&Level{ReturnToState: ParsingEmphasisStrong})
				state = ParsingEmphasisStrong
			case lexeme.Type == lexer.TokenEmphasisStrong:
				_ = levels.Pop()
				paren := levels.Top()
				paren.PushText(gen.EmphasisStrong{
					gen.StringOnlyContent(level.TextValues),
				})
				state = level.ReturnToState
			}
		case ParsingLink:
			level := levels.Top()
			switch {
			default:
				if isTextContent(lexeme.Type) {
					level.PushText(newTextContent(lexeme))
				} else {
					err = errors.Join(err, newError(lexeme, state, errors.New("invalid token")))
				}
			case lexeme.Type == lexer.TokenLinkHref:
				_ = levels.Pop()
				paren := levels.Top()
				paren.PushText(gen.Link{
					Href: lexeme.Text,
					Name: gen.StringOnlyContent(level.TextValues),
				})
				state = level.ReturnToState
			}
		}
	}
	return
}

func isTextContent(tokenType lexer.TokenType) bool {
	switch tokenType {
	default:
		return false
	case lexer.TokenMono, lexer.TokenText, lexer.TokenAmpSpecial:
		return true
	}
	panic("unreachable")
}

func newTextContent(lexeme lexer.Token) gen.StringRenderable {
	Assert(isTextContent(lexeme.Type), fmt.Sprintf("cannot make text content out of %s", lexeme.Type))
	switch lexeme.Type {
	case lexer.TokenMono:
		return gen.Mono(lexeme.Text)
	case lexer.TokenText:
		return gen.Text(lexeme.Text)
	case lexer.TokenAmpSpecial:
		switch lexeme.Text {
		default:
			panic("programmer error: lexer and parser out of sync about what constitutes a &<...>; special character")
		case "~", "\u00A0":
			return gen.NoBreakSpace
		case "---", "&mdash;":
			return gen.EMDash
		case "&ldquo;":
			return gen.LeftDoubleQuote
		case "&rdquo;":
			return gen.RightDoubleQuote
		case "...", "…":
			return gen.Ellipsis
		case "&prime;":
			return gen.Prime
		case "&Prime;":
			return gen.DoublePrime
		case "&tprime;":
			return gen.TripplePrime
		case "&qprime;":
			return gen.QuadruplePrime
		case "&bprime;":
			return gen.ReversedPrime
		}
	}
	panic("unreachable")
}

func checkMetaKey(key string) bool {
	recognizedKeys := map[string]struct{}{
		"author": {},
		"email": {},
		"tags": {},
		//"template": {},
		"title": {},
		"alt-title": {},
		"url-path": {},
		"rel-me": {},
		"fedi-creator": {},
		"lang": {},
		"published": {},
		"revised": {},
		"est-reading": {},
		//"series": {},
		"series-prev": {},
		"series-prev-link": {},
		"series-next": {},
		"series-next-link": {},
		"enable-revision-warning": {},
	}
	_, ok := recognizedKeys[key]
	return ok
}

func setMetaKeyValuePair(blog *gen.Blog, key string, value gen.StringRenderable) (err error) {
	switch key {
	default:
		// do nothing, error already reported
	case "author":
		blog.Author.Name = value
	case "email":
		blog.Author.Email = value
	case "tags":
		for _, tag := range strings.Split(value./*@todo: Clean*/Text(), " ") {
			blog.Tags = append(blog.Tags, gen.Tag(tag))
		}
	case "title":
		blog.Title = value
	case "alt-title":
		blog.AltTitle = value
	case "url-path":
		blog.UrlPath = value./*Clean*/Text()
	case "rel-me":
		blog.Author.RelMe = value
	case "fedi-creator":
		blog.Author.FediCreator = value
	case "lang":
		blog.Lang = value./*Clean*/Text()
	case "published":
		blog.Published.Published, err = time.Parse("2006-01-02", value./*Clean*/Text())
		if err != nil {
			blog.Published.Published, err = time.Parse(time.RFC3339, value./*Clean*/Text())
			if err != nil {
				err = fmt.Errorf("invalid date format, use 2006-01-02 or RFC3339")
			}
		}
	case "revised":
		timeRef := func(t time.Time, err error) (*time.Time, error) {
			return &t, err
		}
		blog.Published.Revised, err = timeRef(time.Parse("2006-01-02", value./*Clean*/Text()))
		if err != nil {
			blog.Published.Revised, err = timeRef(time.Parse(time.RFC3339, value./*Clean*/Text()))
			if err != nil {
				err = fmt.Errorf("invalid date format, use 2006-01-02 or RFC3339")
			}
		}
	case "est-reading":
		blog.EstReading, err = strconv.Atoi(value./*Clean*/Text())
	// @todo: case "series":
	case "series-prev":
		if blog.Series == nil {
			blog.Series = &gen.Series{}
		}
		if blog.Series.Prev == nil {
			blog.Series.Prev = &gen.SeriesItem{}
		}
		blog.Series.Prev.Title = value
	case "series-prev-link":
		if blog.Series == nil {
			blog.Series = &gen.Series{}
		}
		if blog.Series.Prev == nil {
			blog.Series.Prev = &gen.SeriesItem{}
		}
		blog.Series.Prev.Link = value./*Clean*/Text()
	case "series-next":
		if blog.Series == nil {
			blog.Series = &gen.Series{}
		}
		if blog.Series.Next == nil {
			blog.Series.Next = &gen.SeriesItem{}
		}
		blog.Series.Next.Title = value
	case "series-next-link":
		if blog.Series == nil {
			blog.Series = &gen.Series{}
		}
		if blog.Series.Next == nil {
			blog.Series.Next = &gen.SeriesItem{}
		}
		blog.Series.Next.Link = value./*Clean*/Text()
	case "enable-revision-warning":
		switch value.Text() {
		default:
			err = fmt.Errorf("invalid option `%s` for enable-revision-warning, expected one of true, false", value.Text())
		case "true":
			blog.EnableRevisionWarning = true
		case "false":
			blog.EnableRevisionWarning = false
		}
	}
	return
}

type (
	Stack[T any] []T
	Maybe[T any] struct {
		HasValue bool
		Value T
	}
)

func (s Stack[T]) Push(v T) Stack[T] {
	return append(s, v)
}

func (s Stack[T]) Pop() (Stack[T], T) {
	l := len(s)
	Assert(l > 0, "Pop called on empty stack (maybe you want to use SafePop?)")
	return s[:l-1], s[l-1]
}

func (s Stack[T]) SafePop() (Stack[T], Maybe[T]) {
	l := len(s)
	if l > 0 {
		return s[:l-1], Maybe[T]{true, s[l-1]}
	}
	return s, Maybe[T]{HasValue: false}
}

func (s Stack[T]) Peek() T {
	l := len(s)
	return s[l-1]
}

func (s Stack[T]) Empty() (empty Stack[T]) {
	return
}
