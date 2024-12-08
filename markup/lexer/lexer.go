package lexer

import (
	"bytes"
	"errors"
	"fmt"
)

type (
	Lexer struct {
		Filename string
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

func (lx *Lexer) LexSource(filename string, buf *bytes.Buffer) error {
	return errors.New("not implemented")
}

func (t Token) Text() string {
	return "not implemented"
}

func (t Token) String() string {
	return "not implemented"
}
