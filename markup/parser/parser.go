package parser

import (
	"errors"
	"fmt"
	"log"

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

func Parse(lx LexResult) (blog *gen.Blog, err error) {
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
				var (
					metaKey string
					metaVal Maybe[gen.StringRenderable]
				)
				stringValues, metaKey = stringValues.Pop()
				textValues, metaVal = textValues.SafePop()
				setMetaKeyValuePair(blog, metaKey, metaVal)
				stringValues = stringValues.Push(lexeme.Text)
			case lexeme.Type == lexer.TokenMetaEnd:
				var (
					metaKey string
					metaVal Maybe[gen.StringRenderable]
				)
				stringValues, metaKey = stringValues.Pop()
				textValues, metaVal = textValues.SafePop()
				setMetaKeyValuePair(blog, metaKey, metaVal)
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

func setMetaKeyValuePair(blog *gen.Blog, key string, value Maybe[gen.StringRenderable]) {
	switch key {
	default:
	case "author":
	case "email":
	case "tags":
		//case "template":
	case "title":
	case "alt-title":
	case "url-path":
	case "rel-me":
	case "fedi-creator":
	case "lang":
	case "published":
	case "revised":
	case "est-reading":
		//case "series":
	case "series-prev":
	case "series-prev-link":
	case "series-next":
	case "series-next-link":
	case "enable-revision-warning":
	}
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
