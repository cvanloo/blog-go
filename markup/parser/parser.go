package parser

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"strconv"

	. "github.com/cvanloo/blog-go/assert"
	"github.com/cvanloo/blog-go/markup/lexer"
	"github.com/cvanloo/blog-go/markup/gen"
)

type (
	LexResult interface {
		Tokens() func(func(lexer.Token) bool)
	}
	ParserError struct {
		Token lexer.Token
		Inner error
	}
)

func newError(lexeme lexer.Token, inner error) ParserError {
	return ParserError{
		Token: lexeme,
		Inner: inner,
	}
}

func (err ParserError) Error() string {
	return fmt.Sprintf("%s: %s", err.Token.Location(), err.Inner)
}

func Parse(lx LexResult) (blog gen.Blog, err error) {
	type ParseState int
	const (
		ParsingStart ParseState = iota
		ParsingDocument
		ParsingMeta
		ParsingMetaVal
		ParsingHtmlTag
		ParsingSection1
		ParsingSection2
	)
	state := ParsingStart
	var (
		stringValues Stack[string]
		textValues Stack[gen.StringRenderable]
	)
	for lexeme := range lx.Tokens() {
		log.Println(lexeme)
		switch state {
		case ParsingStart:
			switch lexeme.Type {
			default:
				err = errors.Join(err, fmt.Errorf("invalid token, expected one of MetaBegin, HtmlTagOpen, Section1"))
			case lexer.TokenMetaBegin:
				state = ParsingMeta
			case lexer.TokenHtmlTagOpen:
				state = ParsingHtmlTag
			case lexer.TokenSection1:
				state = ParsingSection1
			}
		case ParsingDocument:
		case ParsingMeta:
			switch lexeme.Type {
			default:
				err = errors.Join(err, fmt.Errorf("invalid token, expected one of MetaKey, MetaEnd"))
			case lexer.TokenMetaKey:
				if !checkMetaKey(lexeme.Text) {
					err = errors.Join(err, fmt.Errorf("unrecognized meta key: %s", lexeme.Text))
				}
				stringValues = stringValues.Push(lexeme.Text)
				state = ParsingMetaVal
			case lexer.TokenMetaEnd:
				state = ParsingDocument
			}
		case ParsingMetaVal:
			switch {
			default:
				err = errors.Join(err, fmt.Errorf("invalid token, expected one of MetaKey, MetaEnd, string content"))
			case lexeme.Type == lexer.TokenMetaKey:
				if !checkMetaKey(lexeme.Text) {
					err = errors.Join(err, fmt.Errorf("unrecognized meta key: %s", lexeme.Text))
				}
				var metaKey string
				stringValues, metaKey = stringValues.Pop()
				setMetaKeyValuePair(&blog, metaKey, gen.StringOnlyContent(textValues))
				stringValues = stringValues.Push(lexeme.Text)
				textValues = textValues.Empty()
			case lexeme.Type == lexer.TokenMetaEnd:
				var metaKey string
				stringValues, metaKey = stringValues.Pop()
				setMetaKeyValuePair(&blog, metaKey, gen.StringOnlyContent(textValues))
				textValues = textValues.Empty()
				state = ParsingDocument
			case isStringContent(lexeme.Type):
				textValues = textValues.Push(newTextContent(lexeme))
			}
		case ParsingHtmlTag:
		case ParsingSection1:
		case ParsingSection2:
		}
	}
	return
}

func isStringContent(tokenType lexer.TokenType) bool {
	switch tokenType {
	default:
		return false
	case lexer.TokenMono, lexer.TokenText, lexer.TokenAmpSpecial:
		return true
	}
	panic("unreachable")
}

func newTextContent(lexeme lexer.Token) gen.StringRenderable {
	Assert(isStringContent(lexeme.Type), fmt.Sprintf("cannot make text content out of %s", lexeme.Type))
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
