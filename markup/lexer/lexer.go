package lexer

import (
	"errors"
	"fmt"
	"unicode"
)

type (
	Lexer struct {
		Filename string
		Source string
		Pos int
		Tokens []Token
		Errors []error
	}
	LexerError struct {
		Filename string
		Pos int
		Inner error
	}
	Token struct {
		Type TokenType
		Pos, Len int
	}
)

//go:generate stringer -type TokenType -trimprefix Token
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenMetaStart
	TokenMetaKey
	TokenMetaVal
	TokenMetaEnd
	TokenHtmlTagStart
	TokenHtmlTagAttrKey
	TokenHtmlTagAttrVal
	TokenHtmlTagEnd
	TokenParagraph
	TokenSection1
	TokenSection2
	TokenCodeBlockStart
	TokenCodeBlockEnd
)

func (err LexerError) Error() string {
	return fmt.Sprintf("%s:+%d: %s", err.Filename, err.Pos, err.Inner)
}

func New() *Lexer {
	return &Lexer{}
}

func (lx *Lexer) LexSource(filename, source string) error {
	lx.Filename = filename
	lx.Source = source
	lx.Lex()
	if len(lx.Errors) > 0 {
		return lx.Errors[0]
	}
	return nil
}

func (lx *Lexer) Lex() {
	lx.SkipWhitespace()
	if lx.Pos == 0 && lx.Peek(3) == "---" {
		lx.LexMetaHeader()
	} else if lx.Peek(1) == "#" {
		sectionLevel := 0
		for lx.Peek(1) == "#" {
			sectionLevel += 1
			lx.Next(1)
		}
		if lx.Expect(" ") {
			lx.LexSectionHeader(sectionLevel)
		}
	} else if lx.Peek(1) == "<" {
		lx.LexHtmlTagStart()
	} else if lx.Peek(3) == "```" {
		lx.LexCodeBlockStart()
	} else {
		lx.LexParagraph()
	}
}

func (lx *Lexer) LexMetaHeader() {
	lx.Error(errors.New("LexMetaHeader: not implemented"))
}

func (lx *Lexer) LexSectionHeader(level int) {
	lx.Error(errors.New("LexSectionHeader: not implemented"))
}

func (lx *Lexer) LexHtmlTagStart() {
	lx.Error(errors.New("LexHtmlTagStart: not implemented"))
}

func (lx *Lexer) LexCodeBlockStart() {
	lx.Error(errors.New("LexCodeBlockStart: not implemented"))
}

func (lx *Lexer) LexParagraph() {
	lx.Error(errors.New("LexParagraph: not implemented"))
}

func (lx *Lexer) SkipWhitespace() {
	for unicode.IsSpace(([]rune(lx.Peek(1))[0])) {
		lx.Next(1)
	}
}

func (lx *Lexer) Peek(n int) string {
	return string(lx.Source[lx.Pos:lx.Pos+n])
}

func (lx *Lexer) Next(n int) string {
	start := lx.Pos
	lx.Pos += n
	return string(lx.Source[start:start+n])
}

func (lx *Lexer) Expect(expected string) bool {
	lpos := lx.Pos
	got := lx.Next(len(expected))
	if got != expected {
		lx.ErrorPos(lpos, fmt.Errorf("expected: %s, got: %s", expected, got))
		return false
	}
	return true
}

func (lx *Lexer) ErrorPos(pos int, err error) {
	lx.Errors = append(lx.Errors, LexerError{
		Filename: lx.Filename,
		Pos: pos,
		Inner: err,
	})
}

func (lx *Lexer) Error(err error) {
	lx.Errors = append(lx.Errors, LexerError{
		Filename: lx.Filename,
		Pos: lx.Pos,
		Inner: err,
	})
}

func (t Token) Text() string {
	return "not implemented"
}

func (t Token) String() string {
	return "not implemented"
}
