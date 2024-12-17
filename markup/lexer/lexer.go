package lexer

import (
	"fmt"
	"unicode"
	"strings"
	"errors"

	. "github.com/cvanloo/blog-go/assert"
)

type (
	Lexer struct {
		Filename string
		Source []rune
		Pos, Consumed int
		Lexemes []Token
		Errors []error
	}
	LexerError struct {
		Filename string
		Pos int
		Inner error
	}
	Token struct {
		Type TokenType
		Filename string
		Pos int
		Text string
	}
)

//go:generate stringer -type TokenType -trimprefix Token
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenMetaBegin
	TokenMetaKey
	TokenMetaEnd
	TokenHtmlTagOpen // no paragraphs inside, only TokenText, etc. (but of course can do <p></p> explicitly)
	TokenHtmlTagAttrKey
	TokenHtmlTagAttrVal
	TokenHtmlTagContent
	TokenHtmlTagClose
	TokenParagraphBegin
	TokenParagraphEnd
	TokenSection1Begin
	TokenSection1Content
	TokenSection1End
	TokenSection2Begin
	TokenSection2Content
	TokenSection2End
	TokenMono
	TokenCodeBlockBegin
	TokenCodeBlockLang
	TokenCodeBlockSource
	TokenCodeBlockLineFirst
	TokenCodeBlockLineLast
	TokenCodeBlockEnd
	TokenText
	TokenEmphasis
	TokenStrong
	TokenEmphasisStrong
	TokenLinkHref
	TokenLink
	TokenAmpSpecial
	//TokenEscaped // @todo: https://www.markdownguide.org/basic-syntax/#characters-you-can-escape (handle this on the level of the lexer, not parser -> emit TokenText)
	TokenBlockquoteBegin
	TokenBlockquoteAttrAuthor
	TokenBlockquoteAttrSource
	TokenBlockquoteAttrEnd
	TokenBlockquoteEnd
	TokenEnquoteBegin
	TokenEnquoteEnd
	TokenImageBegin
	TokenImageTitle
	TokenImagePath
	TokenImageAlt
	TokenImageAttrEnd
	TokenImageEnd
	TokenHorizontalRule
	// @todo: sidenote: https://www.markdownguide.org/extended-syntax/#footnotes
)

func (err LexerError) Error() string {
	return fmt.Sprintf("%s:+%d: %s", err.Filename, err.Pos, err.Inner)
}

func New() *Lexer {
	return &Lexer{}
}

func (lx *Lexer) Tokens() func(func(Token) bool) {
	return func(yield func(Token) bool) {
		for _, t := range lx.Lexemes {
			if !yield(t) {
				return
			}
		}
	}
}

// LexSource lexes the passed source and returns the first error that occurred during said lexing, if any.
func (lx *Lexer) LexSource(filename, source string) error {
	// reset lexer state when parsing new file, leave errors though
	lx.Filename = filename
	lx.Source = []rune(source)
	lx.Pos = 0
	lx.Consumed = 0
	firstSourceErrorIdx := len(lx.Errors)
	lx.LexMetaOrContent()
	if len(lx.Errors) > firstSourceErrorIdx {
		return lx.Errors[firstSourceErrorIdx]
	}
	return nil
}

var (
	AmpSpecials = []string{"&mdash;", "&ldquo;", "&rdquo;", "&prime;", "&Prime;", "&tprime;", "&qprime;", "&bprime;"}
)

var (
	SpecAsciiLower = CharInRange{'a', 'z'}
	SpecAsciiUpper = CharInRange{'A', 'Z'}
	SpecAscii = CharInSpec{SpecAsciiLower, SpecAsciiUpper}
	SpecValidMetaKey = CharInSpec{SpecAscii, CharInAny("-_")}
)

type (
	CharSpec interface {
		IsValid(r rune) bool
	}
	CharInRange [2]rune
	CharInAny string
	CharInSpec []CharSpec
)

func (c CharInRange) IsValid(r rune) bool {
	return c[0] <= r && r <= c[1]
}

func (c CharInAny) IsValid(r rune) bool {
	return strings.ContainsRune(string(c), r)
}

func (c CharInSpec) IsValid(r rune) bool {
	for _, spec := range c {
		if spec.IsValid(r) {
			return true
		}
	}
	return false
}

func (lx *Lexer) NextValids(spec CharSpec) string {
	for spec.IsValid(lx.Peek1()) {
		lx.Next(1)
	}
	return lx.Diff()
}

func (lx *Lexer) IsAmpSpecial() bool {
	switch lx.Peek1() {
	case '&':
		for _, special := range AmpSpecials {
			if lx.MatchAtPos(special) {
				return true
			}
		}
	case '…', '~', '\u00A0':
		return true
	}
	switch lx.Peek(3) {
	case "...", "---":
		return true
	}
	return false
}

func (lx *Lexer) LexAmpSpecial() {
	switch lx.Peek1() {
	case '&':
		lx.LexAmpSpecialAmpSemi()
	case '…', '~', '\u00A0':
		lx.Next(1)
		lx.Emit(TokenAmpSpecial)
	default:
		switch lx.Peek(3) {
		case "...", "---":
			lx.Next(3)
			lx.Emit(TokenAmpSpecial)
		default:
			panic("have you checked IsAmpSpecial before calling this function?")
		}
	}
}

func (lx *Lexer) LexAmpSpecialAmpSemi() bool {
	Assert(lx.Peek(1) == "&", "lexer confused")
	//for _, special := range AmpSpecials {
	//	if lx.NextIfMatch(special) {
	//		lx.Emit(TokenAmpSpecial)
	//		return true
	//	}
	//}
	special := lx.NextValids(SpecAscii)
	if !lx.Expect(";") {
		lx.Reset()
		Assert(lx.Peek(1) == "&", "expected to be back at & position after reset")
		return false
	}
	switch special+";" {
	default:
		// invalid
		lx.Error(fmt.Errorf("invalid &<...>; sequence: %s;", special))
		return false
	case "&mdash;", "&ldquo;", "&rdquo;", "&prime;", "&Prime;", "&tprime;", "&qprime;", "&bprime;":
		// valid
		lx.Emit(TokenAmpSpecial)
		return true
	}
}

func (lx *Lexer) LexMetaOrContent() {
	lx.SkipWhitespace()
	if lx.Peek(3) == "---" {
		lx.LexMeta()
	}
	lx.LexContent()
}

func (lx *Lexer) LexMeta() {
	Assert(lx.Peek(3) == "---", "confused lexer state")
	lx.Next(3)
	lx.Emit(TokenMetaBegin)
	if lx.ExpectAndSkip("\n") {
		lx.LexMetaKeyValuePairs()
		lx.Expect("---")
		lx.Emit(TokenMetaEnd)
		lx.ExpectAndSkip("\n")
	}
}

func (lx *Lexer) LexMetaKeyValuePairs() {
	for !lx.IsEOF() && lx.Peek(4) != "---\n" {
		lx.LexMetaKey()
		if lx.ExpectAndSkip(":") {
			lx.LexMetaValue()
		}
		lx.SkipWhitespace()
	}
}

func (lx *Lexer) LexMetaKey() {
	key := lx.NextValids(SpecValidMetaKey)
	lx.Emit(TokenMetaKey)
	_ = key
	if lx.Peek1() != ':' {
		lx.Error(errors.New("a meta key can only contain [a-zA-Z_-]"))
		// try to recover by dropping invalid runes and parsing to the : or \n
		for !lx.IsEOF() && lx.Peek1() != ':' {
			if lx.Peek1() == '\n' {
				// consider key as finished and having an empty value
				lx.Next(1)
				break
			}
			lx.Next(1)
		}
		lx.Skip()
	}
}

func (lx *Lexer) LexMetaValue() {
	lx.SkipWhitespaceNoNewLine()
	if lx.Peek1() != '\n' {
		lx.LexAsStringOrAmpSpecial()
	}
	lx.ExpectAndSkip("\n")
}

func (lx *Lexer) LexAsStringOrAmpSpecial() {
	for !lx.IsEOF() && lx.Peek1() != '\n' {
		if lx.IsAmpSpecial() {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexAmpSpecial()
		} else {
			lx.Next(1)
		}
	}
	lx.EmitIfNonEmpty(TokenText)
}

func (lx *Lexer) LexContent() {
	/*
	lx.SkipWhitespace()
	for !lx.IsEOF() {
		if lx.Peek(3) == "---" {
			lx.Next(3)
			lx.Emit(TokenHorizontalRule)
		} else if lx.Peek(1) == "#" {
			sectionLevel := 0
			for lx.Peek(1) == "#" {
				sectionLevel += 1
				lx.Next(1)
			}
			lx.SkipWhitespace()
			lx.LexSectionHeader(sectionLevel)
		} else if lx.Peek(1) == "<" {
			if lx.Peek(2) == "</" {
				lx.LexHtmlTagEnd()
			} else {
				lx.LexHtmlTagStart()
			}
		} else if lx.Peek(3) == "```" {
			lx.LexCodeBlockStart()
		} else {
			lx.LexParagraph()
		}
		lx.SkipWhitespace()
	}
	lx.Emit(TokenEOF)
	*/
}

func (lx *Lexer) LexMetaKeyValues() {
	lx.SkipWhitespace()
	for !lx.IsEOF() && lx.Peek(3) != "---" {
		key, ok := lx.Until(":")
		if !ok {
			lx.Error(fmt.Errorf("expected key-value pair, got: %s", key))
			break
		}
		lx.Emit(TokenMetaKey)
		// skip past :
		lx.Next(1)
		lx.Skip()
		lx.SkipWhitespace()
		val, ok := lx.Until("\n")
		if !ok {
			lx.Error(fmt.Errorf("expected key-value pair, got: %s", val))
			break
		}
		lx.SkipWhitespace()
	}
}

func (lx *Lexer) LexText() {
	for !lx.IsEOF() {
		if lx.Peek(1) == "~" || lx.Peek(1) == "\u00A0" {
			lx.EmitIfNonEmpty(TokenText)
			lx.Next(1)
			lx.Emit(TokenAmpSpecial)
		} else if lx.Peek(1) == "…" {
			lx.EmitIfNonEmpty(TokenText)
			lx.Next(1)
			lx.Emit(TokenAmpSpecial)
		} else if lx.Peek(3) == "..." {
			lx.EmitIfNonEmpty(TokenText)
			lx.Next(3)
			lx.Emit(TokenAmpSpecial)
		} else if lx.Peek(3) == "---" {
			lx.EmitIfNonEmpty(TokenText)
			lx.Next(3)
			lx.Emit(TokenAmpSpecial)
		} else if lx.Peek(1) == "`" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexMono()
		} else if lx.Peek(1) == "&" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexAmpSpecial()
		} else if lx.Peek(1) == "#" || lx.Peek(1) == "<" || lx.Peek(3) == "```" {
			lx.EmitIfNonEmpty(TokenText)
			break
		} else {
			lx.Next(1)
		}
	}
}

func (lx *Lexer) LexMono() {
	Assert(lx.Peek(1) == "`", "lexer confused")
	// skip past `
	lx.Next(1)
	lx.Skip()
	monoText, closed := lx.Until("`")
	_ = monoText
	if !closed {
		// @todo ???
		lx.Error(fmt.Errorf("unclosed `"))
	}
	lx.Emit(TokenMono)
	lx.Expect("`")
}

func (lx *Lexer) LexSectionHeader(level int) {
	lx.Until("\n")
	switch level {
	default:
		lx.Error(fmt.Errorf("invalid section level: %d", level))
	case 1:
		lx.Emit(TokenSection1Begin)
	case 2:
		lx.Emit(TokenSection2Begin)
	}
}

func (lx *Lexer) LexHtmlTagStart() {
	// skip past <
	lx.Next(1)
	lx.Skip()
	if tagName := lx.NextASCII(); tagName == "" {
		lx.Error(fmt.Errorf("expected html tag name"))
	}
	lx.Emit(TokenHtmlTagOpen)
	lx.SkipWhitespace()
	if lx.Peek(1) == ">" {
		lx.Next(1)
		lx.Skip()
	} else {
		lx.LexHtmlTagAttrs()
		lx.Expect(">")
		lx.Skip()
	}
}

func (lx *Lexer) LexHtmlTagAttrs() {
	lx.SkipWhitespace()
	for lx.Peek(1) != ">" {
		if attrKey := lx.NextASCII(); attrKey == "" {
			lx.Error(fmt.Errorf("expected attribute or >, got: %s", lx.Peek(1)))
			break
		}
		lx.Emit(TokenHtmlTagAttrKey)
		lx.SkipWhitespace()
		if lx.Peek(1) == "=" {
			lx.Next(1)
			lx.SkipWhitespace()
			lx.Expect("\"")
			lx.Skip()
			val, ok := lx.Until("\"")
			if !ok {
				lx.Error(fmt.Errorf("expected value delimited by double quotes, got: %s", val))
			}
			lx.Emit(TokenHtmlTagAttrVal)
			// skip past "
			lx.Next(1)
			lx.Skip()
		}
		lx.SkipWhitespace()
	}
}

func (lx *Lexer) LexHtmlTagEnd() {
	// skip past </
	lx.Next(2)
	lx.Skip()
	if tagName := lx.NextASCII(); tagName == "" {
		lx.Error(fmt.Errorf("expected html tag name"))
	}
	lx.Emit(TokenHtmlTagClose)
	lx.SkipWhitespace()
	lx.Expect(">")
	lx.Skip()
}

func (lx *Lexer) LexParagraph() {
	lx.Skip()
	lx.Emit(TokenParagraphBegin)
	for !lx.IsEOF() {
		if lx.Peek(1) == "<" {
			lx.Emit(TokenText)
			if lx.Peek(2) == "</" {
				lx.LexHtmlTagEnd()
			} else {
				lx.LexHtmlTagStart()
			}
			continue
		}
		if lx.Peek(1) == "#" || lx.Peek(1) == "\n\n" || lx.Peek(3) == "```" {
			lx.Emit(TokenText)
			lx.Emit(TokenParagraphEnd)
			break
		}
		lx.Next(1)
	}
}

func (lx *Lexer) LexCodeBlockStart() {
	// skip past ```
	lx.Next(3)
	lx.Skip()
	lx.NextASCII()
	lx.Emit(TokenCodeBlockBegin)
	lx.SkipWhitespace()
	lx.Until("```")
	lx.Emit(TokenText)
	lx.Emit(TokenCodeBlockEnd)
	lx.Next(3)
	lx.Skip()
}

func (lx *Lexer) SkipWhitespace() {
	for !lx.IsEOF() && unicode.IsSpace(([]rune(lx.Peek(1))[0])) {
		lx.Next(1)
	}
	lx.Skip()
}

func (lx *Lexer) SkipWhitespaceNoNewLine() {
	for !lx.IsEOF() && unicode.IsSpace(lx.Peek1()) && lx.Peek1() != '\n' {
		lx.Next(1)
	}
	lx.Skip()
}

func (lx *Lexer) IsEOF() bool {
	return lx.Pos >= len(lx.Source)
}

func (lx *Lexer) Diff() string {
	return string(lx.Source[lx.Consumed:lx.Pos])
}

func (lx *Lexer) Peek1() rune {
	if lx.IsEOF() {
		return 0
	}
	return lx.Source[lx.Pos]
}

func (lx *Lexer) Peek(n int) string {
	if lx.IsEOF() {
		return ""
	}
	m := min(lx.Pos+n, len(lx.Source))
	return string(lx.Source[lx.Pos:m])
}

func (lx *Lexer) Next(n int) string {
	if lx.IsEOF() {
		return ""
	}
	start := lx.Pos
	m := min(n, len(lx.Source) - start)
	lx.Pos += m
	return string(lx.Source[start:start+m])
}

func (lx *Lexer) Until(search string) (string, bool) {
	lpos := lx.Pos
	for {
		if len(lx.Source) - lx.Pos >= len(search) {
			if lx.Peek(len(search)) == search {
				return string(lx.Source[lpos:lx.Pos]), true
			} else {
				lx.Next(1)
			}
		} else {
			return string(lx.Source[lpos:lx.Pos]), false
		}
	}
}

func (lx *Lexer) NextASCII() string {
	lpos := lx.Pos
	for !lx.IsEOF() {
		r := []rune(lx.Peek(1))[0]
		if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')) {
			break
		}
		lx.Next(1)
	}
	return string(lx.Source[lpos:lx.Pos])
}

func (lx *Lexer) Skip() {
	lx.Consumed = lx.Pos
}

func (lx *Lexer) Reset() {
	lx.Pos = lx.Consumed
}

func (lx *Lexer) Emit(tokenType TokenType) {
	lx.Lexemes = append(lx.Lexemes, Token{
		Filename: lx.Filename,
		Type: tokenType,
		Pos: lx.Consumed,
		Text: string(lx.Source[lx.Consumed:lx.Pos]),
	})
	lx.Consumed = lx.Pos
}

func (lx *Lexer) EmitIfNonEmpty(tokenType TokenType) {
	if lx.Pos > lx.Consumed {
		lx.Emit(tokenType)
	}
}

func (lx *Lexer) MatchAtPos(test string) bool {
	got := lx.Peek(len(test))
	return got == test
}

func (lx *Lexer) NextIfMatch(test string) bool {
	got := lx.Peek(len(test))
	if got != test {
		return false
	}
	lx.Next(len(test))
	return true
}

func (lx *Lexer) Expect(expected string) bool {
	got := lx.Peek(len(expected))
	if got != expected {
		lx.Error(fmt.Errorf("expected: %s, got: %s", expected, got))
		return false
	}
	lx.Next(len(expected))
	return true
}

func (lx *Lexer) ExpectAndSkip(expected string) bool {
	got := lx.Peek(len(expected))
	if got != expected {
		lx.Error(fmt.Errorf("expected: %s, got: %s", expected, got))
		return false
	}
	lx.Next(len(expected))
	lx.Skip()
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

func (t Token) String() string {
	return fmt.Sprintf("%s:+%d: %s: `%s`", t.Filename, t.Pos, t.Type, t.Text)
}

func (t Token) Location() string {
	return fmt.Sprintf("%s:+%d", t.Filename, t.Pos)
}
