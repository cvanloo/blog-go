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
		Filename string `deep:"-"`
		Pos int `deep:"-"`
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
	TokenHtmlTagOpen // ??? no paragraphs inside, only TokenText, etc. (but of course can do <p></p> explicitly)
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
	TokenLinkText
	TokenLinkHref
	TokenLinkIdRef
	TokenLinkIdDef
	TokenAmpSpecial
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
	TokenSidenote
)

func (t Token) String() string {
	return fmt.Sprintf("%s:+%d: %s: `%s`", t.Filename, t.Pos, t.Type, t.Text)
}

func (t Token) Location() string {
	return fmt.Sprintf("%s:+%d", t.Filename, t.Pos)
}

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

func (lx *Lexer) Clear() {
	lx.Lexemes = nil
	lx.Errors = nil
}

func (lx *Lexer) Skip() {
	lx.Consumed = lx.Pos
}

func (lx *Lexer) Reset() {
	lx.Pos = lx.Consumed
}

func (lx *Lexer) ResetToPos(pos int) {
	lx.Pos = pos
}

func (lx *Lexer) DeferResetToPos(pos int) func() {
	return func() {
		lx.ResetToPos(pos)
	}
}

func (lx *Lexer) DeferResetToCurrentPos() func() {
	pos := lx.Pos
	return func() {
		lx.ResetToPos(pos)
	}
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

func (lx *Lexer) NextMatchSurroundedByNewlines(search string) bool {
	lx.SkipWhitespaceNoNewLine()
	if !lx.MatchAtPos("\n") {
		return false
	}
	lx.Next1()
	lx.SkipWhitespace()
	if !lx.MatchAtPos(search) {
		return false
	}
	lx.SkipWhitespaceNoNewLine()
	if !lx.MatchAtPos("\n") {
		return false
	}
	lx.SkipWhitespace()
	return true
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

func (lx *Lexer) Next1() rune {
	if lx.IsEOF() {
		return 0
	}
	defer func() {
		lx.Pos++
	}()
	return lx.Source[lx.Pos]
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

func (lx *Lexer) NextValids(spec CharSpec) string {
	for spec.IsValid(lx.Peek1()) {
		lx.Next(1)
	}
	return lx.Diff()
}

func (lx *Lexer) UntilSpec(spec CharSpec) string {
	for !spec.IsValid(lx.Peek1()) {
		lx.Next1()
	}
	return lx.Diff()
}

func (lx *Lexer) UntilMatch(search string) (string, bool) {
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

func (lx *Lexer) IsStartOfLine() bool {
	return lx.Pos == 0 || lx.Source[lx.Pos-1] == '\n'
}

func (lx *Lexer) NextEmptyLine() bool {
	lx.SkipWhitespaceNoNewLine()
	return lx.Peek1() == '\n'
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

func (lx *Lexer) EmitIfNonEmpty(tokenType TokenType) bool {
	if lx.Pos > lx.Consumed {
		lx.Emit(tokenType)
		return true
	}
	return false
}

func (lx *Lexer) MatchAtPos(test string) bool {
	got := lx.Peek(len(test))
	return got == test
}

func (lx *Lexer) NextIfMatch(test string) bool {
	if !lx.MatchAtPos(test) {
		return false
	}
	lx.Next(len(test))
	return true
}

func (lx *Lexer) Expect(expected string) bool {
	got := lx.Peek(len(expected))
	if got != expected {
		lx.Error(fmt.Errorf("expected: `%s`, got: `%s`", WhiteSpaceToVisible(expected), WhiteSpaceToVisible(got)))
		return false
	}
	lx.Next(len(expected))
	return true
}

func (lx *Lexer) ExpectAndSkip(expected string) bool {
	got := lx.Peek(len(expected))
	if got != expected {
		lx.Error(fmt.Errorf("expected: `%s`, got: `%s`", WhiteSpaceToVisible(expected), WhiteSpaceToVisible(got)))
		return false
	}
	lx.Next(len(expected))
	lx.Skip()
	return true
}

func (lx *Lexer) Error(err error) {
	lx.Errors = append(lx.Errors, LexerError{
		Filename: lx.Filename,
		Pos: lx.Pos,
		Inner: err,
	})
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
	AmpShortSpecials = []string{"…", "...", "~", "\u00A0", "---"}
	AmpAllSpecials = append(AmpSpecials, AmpShortSpecials...)
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
	CharNotInSpec []CharSpec
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

func (c CharNotInSpec) IsValid(r rune) bool {
	for _, spec := range c {
		if spec.IsValid(r) {
			return false
		}
	}
	return true
}

func (lx *Lexer) IsAmpSpecial() (bool, string) {
	for _, special := range AmpAllSpecials {
		if lx.MatchAtPos(special) {
			return true, special
		}
	}
	return false, ""
}

func (lx *Lexer) LexAmpSpecial() bool {
	if ok, special := lx.IsAmpSpecial(); ok {
		lx.Next(len(special))
		lx.Emit(TokenAmpSpecial)
		return true
	}
	// just for showing the user a nice error
	if lx.Peek1() == '&' {
		special := lx.NextValids(SpecAscii)
		if lx.Peek1() == ';' {
			lx.Error(fmt.Errorf("invalid &<...>; sequence: %s;", special))
		}
		lx.Reset()
	}
	// & don't have to be escaped, if it isn't of the form &[a-zA-Z];
	return false

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
	lx.SkipWhitespaceNoNewLine()
	if lx.ExpectAndSkip("\n") {
		lx.LexMetaKeyValuePairs()
		lx.Expect("---")
		lx.Emit(TokenMetaEnd)
		lx.SkipWhitespaceNoNewLine()
		lx.ExpectAndSkip("\n")
	}
}

func (lx *Lexer) LexMetaKeyValuePairs() {
	lx.SkipWhitespace()
	for !lx.IsEOF() && lx.Peek(3) != "---" {
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
		if ok, _ := lx.IsAmpSpecial(); ok {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexAmpSpecial()
		} else {
			lx.Next(1)
		}
	}
	lx.EmitIfNonEmpty(TokenText)
}

func (lx *Lexer) LexContent() {
	for !lx.IsEOF() {
		if lx.NextMatchSurroundedByNewlines("---") || lx.NextMatchSurroundedByNewlines("***") {
			lx.Emit(TokenHorizontalRule)
		} else if lx.IsDefinition() {
			lx.LexDefinition()
		} else {
			lx.SkipWhitespace()
			if lx.Peek(3) == "```" {
				lx.LexCodeBlock()
			} else if lx.Peek(3) == "***" || lx.Peek(3) == "___" {
				lx.LexEmphasisStrong()
			} else if lx.Peek(2) == "**" || lx.Peek(2) == "__" {
				lx.LexStrong()
			} else if lx.Peek(2) == "/>" {
				lx.LexHtmlTagClose()
			} else if lx.Peek(2) == "![" {
				lx.LexImage()
			} else if lx.Peek(2) == "~~" {
				lx.LexStrikethrough()
			} else if lx.Peek(2) == "==" {
				lx.LexMarker()
			} else if lx.Peek1() == '*' || lx.Peek1() == '_' {
				lx.LexEmphasis()
			} else if lx.Peek1() == '<' {
				lx.LexLinkOrHtmlTagOpen()
			} else if lx.Peek1() == '>' {
				lx.LexBlockQuote()
			} else if lx.Peek1() == '#' {
				lx.LexSection()
			} else if lx.Peek1() == '`' {
				lx.LexMono()
			} else if lx.Peek1() == '[' {
				lx.LexLinkOrSidenote()
			} else if lx.Peek1() == '"' {
				lx.LexEnquote()
			} else if lx.Peek1() == `\` {
				lx.LexEscape()
			}
		}
	}
	lx.Emit(TokenEOF)
}

func (lx *Lexer) LexSection() {
	Assert(lx.Peek1() == '#', "lexer state confused")
	hashes := lx.NextValids(CharInAny("#"))
	level := len(hashes)
	switch level {
	case 1:
		lx.Emit(TokenSection1Begin)
	case 2:
		lx.Emit(TokenSection2Begin)
	default:
		lx.Error(errors.New("max section level is 2"))
	}
	lx.ExpectAndSkip(" ")
	heading := lx.UntilSpec(CharInAny("\n"))
	if len(heading) == 0 {
		lx.Error(errors.New("section must have a heading"))
	}
	switch level {
	case 1:
		lx.Emit(TokenSection1Content)
	case 2:
		lx.Emit(TokenSection2Content)
	}
}

func (lx *Lexer) IsDefinition() bool {
	if !lx.IsStartOfLine() {
		return false
	}
	lx.NextValids(SpecValidDefinitionTerm)
	if lx.Peek1() != '\n' {
		return false
	}
	lx.Next1()
	lx.SkipWhitespaceNoNewLine()
	if lx.Peek1() != ':' {
		return false
	}
	lx.Reset()
	return true
}

/*
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
*/

func WhiteSpaceToVisible(s string) string {
	var builder strings.Builder
	for _, r := range s {
		switch r {
		default:
			builder.WriteRune(r)
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		case '\v':
			builder.WriteString(`\v`)
		}
	}
	return builder.String()
}
