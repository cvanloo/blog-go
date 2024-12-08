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
		Pos, Consumed int
		Tokens []Token
		Errors []error
	}
	LexerError struct {
		Filename string
		Pos int
		Inner error
	}
	Token struct {
		lx *Lexer
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
	lx.LexStart()
	if len(lx.Errors) > 0 {
		return lx.Errors[0]
	}
	return nil
}

func (lx *Lexer) LexStart() {
	defer func() {
		r := recover()
		lx.Error(fmt.Errorf("%v", r))
	}()
	lx.SkipWhitespace()
	if lx.Peek(3) == "---" {
		lx.Next(3)
		lx.Emit(TokenMetaStart)
		lx.LexMetaKeyValues()
		if lx.Expect("---") {
			lx.Emit(TokenMetaEnd)
		}
	}
	lx.LexContent()
}

func (lx *Lexer) LexContent() {
	for !lx.IsEOF() {
		lx.SkipWhitespace()
		/*if lx.Peek(3) == "---" {
			lx.Next(3)
			lx.LexHorizontalRuler()
		} else */
		if lx.Peek(1) == "#" {
			sectionLevel := 0
			for lx.Peek(1) == "#" {
				sectionLevel += 1
				lx.Next(1)
			}
			if lx.Expect(" ") {
				lx.Skip()
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
}

func (lx *Lexer) LexMetaKeyValues() {
	lx.Error(errors.New("LexMetaKeyValues: not implemented"))
	for lx.Peek(3) != "---" {
		_, ok := lx.Until(":")
		if !ok {
			break
		}
		lx.Skip()
		lx.Next(1)
		lx.Skip()
	}
}

func (lx *Lexer) LexSectionHeader(level int) {
	lx.Until("\n")
	switch level {
	default:
		lx.Error(fmt.Errorf("invalid section level: %d", level))
	case 1:
		lx.Emit(TokenSection1)
	case 2:
		lx.Emit(TokenSection2)
	}
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

func (lx *Lexer) IsEOF() bool {
	return lx.Pos >= len(lx.Source)
}

func (lx *Lexer) Peek(n int) string {
	return string(lx.Source[lx.Pos:lx.Pos+n])
}

func (lx *Lexer) Next(n int) string {
	start := lx.Pos
	lx.Pos += n
	return string(lx.Source[start:start+n])
}

func (lx *Lexer) Until(search string) (string, bool) {
	lpos := lx.Pos
	for {
		if len(lx.Source) - lx.Pos >= len(search) {
			if lx.Peek(len(search)) == search {
				return lx.Source[lpos:lx.Pos], true
			} else {
				lx.Next(1)
			}
		} else {
			return lx.Source[lpos:lx.Pos], false
		}
	}
}

func (lx *Lexer) Skip() {
	lx.Consumed = lx.Pos
}

func (lx *Lexer) Emit(tokenType TokenType) {
	lx.Tokens = append(lx.Tokens, Token{
		lx: lx,
		Type: tokenType,
		Pos: lx.Consumed,
		Len: lx.Pos - lx.Consumed,
	})
	lx.Consumed = lx.Pos
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
	return t.lx.Source[t.Pos:t.Pos+t.Len]
}

func (t Token) String() string {
	return fmt.Sprintf("%s:+%d: %s: %s", t.lx.Filename, t.Pos, t.Type, t.Text())
}
