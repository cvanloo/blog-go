package lexer

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	. "github.com/cvanloo/blog-go/assert"
)

type (
	Lexer struct {
		Filename      string
		Source        []rune
		Pos, Consumed int
		Lexemes       []Token
		Errors        []error
	}
	LexerError struct {
		Filename string
		Pos      int
		Inner    error
	}
	Token struct {
		Type     TokenType
		Filename string `deep:"-"`
		Pos      int    `deep:"-"`
		Text     string
	}
)

//go:generate stringer -type TokenType -trimprefix Token
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenMetaBegin
	TokenMetaKey
	TokenMetaEnd
	TokenHtmlTagOpen
	TokenHtmlTagAttrKey
	TokenHtmlTagAttrVal
	TokenHtmlTagContent
	TokenHtmlTagClose
	TokenSection1Begin
	TokenSection1Content
	TokenSection1End
	TokenSection2Begin
	TokenSection2Content
	TokenSection2End
	TokenParagraphBegin
	TokenParagraphEnd
	TokenText
	TokenAmpSpecial
	TokenMono
	TokenEmphasisBegin
	TokenEmphasisEnd
	TokenStrikethroughBegin
	TokenStrikethroughEnd
	TokenMarkerBegin
	TokenMarkerEnd
	TokenStrongBegin
	TokenStrongEnd
	TokenEmphasisStrongBegin
	TokenEmphasisStrongEnd
	TokenEnquoteSingleBegin
	TokenEnquoteSingleEnd
	TokenEnquoteDoubleBegin
	TokenEnquoteDoubleEnd
	TokenEnquoteAngledBegin
	TokenEnquoteAngledEnd
	TokenDefinitionTerm
	TokenDefinitionExplanationBegin
	TokenDefinitionExplanationEnd
	TokenHorizontalRule
	TokenBlockquoteBegin
	TokenBlockquoteAttrAuthor
	TokenBlockquoteAttrSource
	TokenBlockquoteAttrEnd
	TokenBlockquoteEnd
	TokenImageBegin
	TokenImageAltText
	TokenImagePath
	TokenImageTitle
	TokenImageEnd
	TokenSidenoteRef
	TokenSidenoteDef
	TokenSidenoteDefEnd
	TokenSidenoteContent
	TokenLinkify
	TokenLinkHref
	TokenLinkRef
	TokenLinkDef
	TokenLinkableBegin
	TokenLinkableEnd
	TokenCodeBlockBegin
	TokenCodeBlockLang
	TokenCodeBlockEnd
	TokenAttributeListBegin
	TokenAttributeListID
	TokenAttributeListKey
	TokenAttributeListEnd
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

func (lx *Lexer) Reset() {
	lx.Pos = lx.Consumed
}

func (lx *Lexer) ResetToPos(pos int) {
	lx.Pos = pos
}

func (lx *Lexer) IsEOF() bool {
	return lx.Pos >= len(lx.Source)
}

func (lx *Lexer) IsStartOfLine() bool {
	return lx.Pos == 0 || lx.Source[lx.Pos-1] == '\n'
}

func (lx *Lexer) IsEmphasis() bool {
	return lx.Peek1() == '*' || lx.Peek1() == '_'
}

func (lx *Lexer) IsStrong() bool {
	return lx.Peek(2) == "**" || lx.Peek(2) == "__"
}

func (lx *Lexer) IsEmphasisStrong() bool {
	return lx.Peek(3) == "***" || lx.Peek(3) == "___"
}

func (lx *Lexer) IsEscape() bool {
	lpos := lx.Pos
	defer lx.ResetToPos(lpos)
	if lx.Peek1() != '\\' {
		return false
	}
	lx.Next1()
	switch lx.Peek1() {
	default:
		return false
	case '\\', '!', '`', '*', '_', '{', '}', '<', '>', '[', ']', '(', ')', '|', '#', '+', '-', '.':
		return true
	case '&':
		ok, _ := lx.IsAmpSpecial()
		return ok
	}
}

type Predicate func() bool

var (
	AmpSpecials      = []string{"&hyphen;", "&dash;", "&ndash;", "&mdash;", "&ldquo;", "&rdquo;", "&prime;", "&Prime;", "&tprime;", "&qprime;", "&bprime;", "&laquo;", "&raquo;", "&nbsp;"}
	AmpShortSpecials = []string{"---", "--", "...", "…", "~", "\u00A0"} // list longer char sequences first (--- must come before -- and -)
	AmpAllSpecials   = append(AmpSpecials, AmpShortSpecials...)
)

func (lx *Lexer) IsAmpSpecial() (bool, string) {
	for _, special := range AmpAllSpecials {
		if lx.MatchAtPos(special) {
			if special == "~" && lx.MatchAtPos("~~") { // @todo: a bit ugly, isn't it?
				continue
			}
			return true, special
		}
	}
	return false, ""
}

func (lx *Lexer) IsParagraphEnder() bool {
	switch {
	default:
		return false
	case lx.IsTermDefinition():
		fallthrough
	case lx.IsLinkOrSidenoteDefinition():
		fallthrough
	case lx.Peek(3) == "```":
		fallthrough
	case lx.Peek(2) == "![":
		fallthrough
	case lx.Peek(2) == "</":
		fallthrough
	case lx.Peek(2) == "\n\n":
		fallthrough
	//case lx.Peek1() == '[':
	//	fallthrough
	case lx.Peek1() == '>':
		return true
	}
}

func (lx *Lexer) IsTermDefinition() bool {
	lpos := lx.Pos
	lcon := lx.Consumed
	defer func() {
		lx.Pos = lpos
		lx.Consumed = lcon
	}()
	lx.NextValids(SpecNonWhitespace)
	lx.SkipWhitespaceNoNewLine()
	if lx.Peek1() != '\n' {
		return false
	}
	lx.SkipNext1()
	lx.SkipWhitespaceNoNewLine()
	if lx.Peek1() != ':' {
		return false
	}
	return true
}

func (lx *Lexer) IsSingleWordSidenote() bool {
	lpos := lx.Pos
	lcon := lx.Consumed
	defer func() {
		lx.Pos = lpos
		lx.Consumed = lcon
	}()
	if !SpecNonWhitespace.IsValid(lx.Peek1()) {
		return false
	}
	lx.NextValids(SpecSingleWordSidenote)
	if lx.Peek(2) != "[^" && lx.Peek(2) != "(^" {
		return false
	}
	return true
}

func (lx *Lexer) IsLinkOrSidenoteDefinition() bool {
	lpos := lx.Pos
	lcon := lx.Consumed
	defer func() {
		lx.Pos = lpos
		lx.Consumed = lcon
	}()
	lx.SkipWhitespaceNoNewLine()
	if lx.Peek1() != '[' {
		return false
	}
	lx.SkipNext1()
	if lx.Peek1() == '^' {
		lx.SkipNext1()
	}
	lx.NextUntilMatch("]")
	if lx.Peek1() != ']' {
		return false
	}
	lx.SkipNext1()
	if lx.Peek1() != ':' {
		return false
	}
	return true
}

func (lx *Lexer) IsHorizontalRule() bool {
	lpos := lx.Pos
	lcon := lx.Consumed
	defer func() {
		lx.Pos = lpos
		lx.Consumed = lcon
	}()
	lx.SkipWhitespaceNoNewLine()
	if lx.Peek1() != '\n' {
		return false
	}
	lx.SkipNext1()
	lx.SkipWhitespace()
	if !(lx.Peek(3) == "---" || lx.Peek(3) == "***") {
		return false
	}
	lx.SkipNext(3)
	lx.SkipWhitespaceNoNewLine()
	if lx.Peek1() != '\n' {
		return false
	}
	return true
}

func (lx *Lexer) Diff() string {
	return string(lx.Source[lx.Consumed:lx.Pos])
}

func (lx *Lexer) Skip() {
	lx.Consumed = lx.Pos
}

func (lx *Lexer) SkipWhitespace() {
	for !lx.IsEOF() && unicode.IsSpace(lx.Peek1()) {
		lx.Next1()
	}
	lx.Skip()
}

func (lx *Lexer) SkipWhitespaceNoNewLine() {
	for !lx.IsEOF() && unicode.IsSpace(lx.Peek1()) && lx.Peek1() != '\n' {
		lx.Next1()
	}
	lx.Skip()
}

func (lx *Lexer) SkipNext(n int) {
	lx.Next(n)
	lx.Skip()
}

func (lx *Lexer) SkipNext1() {
	lx.SkipNext(1)
}

func (lx *Lexer) Peek(n int) string {
	if lx.IsEOF() {
		return ""
	}
	m := min(lx.Pos+n, len(lx.Source))
	return string(lx.Source[lx.Pos:m])
}

func (lx *Lexer) Peek1() rune {
	if lx.IsEOF() {
		return 0
	}
	return lx.Source[lx.Pos]
}

func (lx *Lexer) MatchAtPos(test string) bool {
	got := lx.Peek(len(test))
	return got == test
}

func (lx *Lexer) Next(n int) string {
	if lx.IsEOF() {
		return ""
	}
	start := lx.Pos
	m := min(n, len(lx.Source)-start)
	lx.Pos += m
	return string(lx.Source[start : start+m])
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

func (lx *Lexer) NextValids(spec CharSpec) string {
	for spec.IsValid(lx.Peek1()) {
		lx.Next1()
	}
	return lx.Diff()
}

func (lx *Lexer) NextUntilSpec(spec CharSpec) (string, bool) {
	for !lx.IsEOF() && !spec.IsValid(lx.Peek1()) {
		lx.Next1()
	}
	return lx.Diff(), spec.IsValid(lx.Peek1())
}

func (lx *Lexer) NextUntilMatch(search string) (string, bool) {
	lpos := lx.Pos
	for {
		if len(lx.Source)-lx.Pos >= len(search) {
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

func (lx *Lexer) NextIfMatch(test string) bool {
	if !lx.MatchAtPos(test) {
		return false
	}
	lx.Next(len(test))
	return true
}

func (lx *Lexer) NextEmptyLine() bool {
	lx.SkipWhitespaceNoNewLine()
	return lx.Peek1() == '\n'
}

func (lx *Lexer) Emit(tokenType TokenType) {
	lx.Lexemes = append(lx.Lexemes, Token{
		Filename: lx.Filename,
		Type:     tokenType,
		Pos:      lx.Consumed,
		Text:     string(lx.Source[lx.Consumed:lx.Pos]),
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
		Pos:      lx.Pos,
		Inner:    err,
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

type (
	CharSpec interface {
		IsValid(r rune) bool
	}
	CharInRange   [2]rune
	CharInAny     string
	CharInSpec    []CharSpec
	CharNotInSpec []CharSpec
)

var (
	SpecAsciiLower         = CharInRange{'a', 'z'}
	SpecAsciiUpper         = CharInRange{'A', 'Z'}
	SpecAscii              = CharInSpec{SpecAsciiLower, SpecAsciiUpper}
	SpecNumber             = CharInRange{'0', '9'}
	SpecAsciiID            = CharInSpec{SpecAscii, SpecNumber, CharInAny("-_")}
	SpecValidMetaKey       = CharInSpec{SpecAscii, CharInAny("-_")}
	SpecNonWhitespace      = CharNotInSpec{CharInAny(" \u00A0\n\r\v\t")} // @todo: and all the others...
	SpecAttrVal            = CharNotInSpec{CharInAny(" \u00A0\n\r\v\t}")}
	SpecAttrKey            = CharNotInSpec{CharInAny(" \u00A0\n\r\v\t=}")}
	SpecImagePath          = CharNotInSpec{CharInAny(" \u00A0\n\r\v\t)")}
	SpecSingleWordSidenote = CharNotInSpec{CharInAny(" \u00A0\n\r\v\t[(")}
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
	return r != 0 // 0 is never valid
}

// LexAmpSpecial lexes a special character, that is one of the &<...>; sequences from AmpSecials
// or from the shortcuts in AmpShortSpecials.
// (They are all defined together in AmpAllSpecials.)
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

// LexMetaOrContent lexes an optional meta data block of the form
//
//	---
//	key: value
//	other_key: some longer string value with spaces
//	---
//
// Instead of `---` it is also possible to use `+++`.
// And instead of `key: value` it is also okay to use `key = value`.
//
// - TokenMetaBegin "---"
// - TokenMetaKey "key"
// - TokenText "value"
// - TokenMetaKey "other_key"
// - TokenText "some longer string value with spaces"
// - TokenMetaEnd "---"
//
// After the meta block is lexed, LexMetaOrContent goes over to lexing the content.
func (lx *Lexer) LexMetaOrContent() {
	lx.SkipWhitespace()
	if lx.Peek(3) == "---" || lx.Peek(3) == "+++" {
		lx.LexMetaWithDelimiter(lx.Peek(3))
	}
	lx.LexContent()
}

func (lx *Lexer) LexMetaWithDelimiter(metaDelim string) {
	Assert(lx.MatchAtPos(metaDelim), "confused lexer state")
	lx.Next(len(metaDelim))
	lx.Emit(TokenMetaBegin)
	lx.SkipWhitespaceNoNewLine()
	if lx.ExpectAndSkip("\n") {
		lx.LexMetaKeyValuePairs()
		if lx.Peek(3) != metaDelim {
			lx.Error(fmt.Errorf("expected: `%s`, got: `%s`", metaDelim, lx.Peek(3)))
			if len(lx.Peek(3)) == 3 {
				lx.Next(3)
			}
		} else {
			lx.Next(3)
		}
		lx.Emit(TokenMetaEnd)
		lx.SkipWhitespaceNoNewLine()
		lx.ExpectAndSkip("\n")
	}
}

func (lx *Lexer) LexMetaKeyValuePairs() {
	lx.SkipWhitespace()
	for !lx.IsEOF() && !lx.MatchAtPos("+++") && !lx.MatchAtPos("---") {
		lx.LexMetaKey()
		if lx.Peek1() == ':' || lx.Peek1() == '=' {
			lx.SkipNext1()
			lx.LexMetaValue()
		} else {
			lx.Error(errors.New("expected : or = after key and before value"))
		}
		lx.SkipWhitespace()
	}
}

func (lx *Lexer) LexMetaKey() {
	key := lx.NextValids(SpecValidMetaKey)
	_ = key
	lx.Emit(TokenMetaKey)
	lx.SkipWhitespaceNoNewLine()
	if !(lx.Peek1() == ':' || lx.Peek1() == '=') {
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
	for !lx.IsEOF() {
		lx.SkipWhitespaceNoNewLine()
		if lx.Peek(2) == "\\\n" {
			lx.SkipNext(2)
		} else if lx.Peek1() != '\n' {
			lx.LexAsMultiLineStringOrAmpSpecial()
			lx.ExpectAndSkip("\n")
			break
		}
	}
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

func (lx *Lexer) LexAsMultiLineStringOrAmpSpecial() {
	for !lx.IsEOF() {
		if lx.Peek(2) == "\\\n" {
			lx.EmitIfNonEmpty(TokenText)
			lx.SkipNext(2)
			lx.SkipWhitespaceNoNewLine()
			lx.Next1()
		} else if lx.Peek1() == '\n' {
			break
		} else {
			if ok, _ := lx.IsAmpSpecial(); ok {
				lx.EmitIfNonEmpty(TokenText)
				lx.LexAmpSpecial()
			} else {
				lx.Next(1)
			}
		}
	}
	lx.EmitIfNonEmpty(TokenText)
}

// LexContent lexes the top level of a markdown document (after the meta block).
// At this level only section 1, html tags and term, link, and sidenote definitions are allowed.
func (lx *Lexer) LexContent() {
	for !lx.IsEOF() {
		lx.SkipWhitespace()
		if lx.Peek1() == '#' {
			lx.LexSection1()
		} else if lx.Peek1() == '<' {
			lx.LexHtmlElement()
		} else if lx.Peek1() == '[' {
			lx.LexLinkOrSidenoteDefinition()
		}
		// @todo: term definitions
		// @todo: invalid stuff... (don't get stuck here forever)
	}
	lx.Emit(TokenEOF)
}

// LexSection1 lexes a top-level section.
//
//	# Section 1
//
// - TokenSection1Begin "#"
// - TokenText "Section 1"
// - TokenSection1Content
// - <...>
// - TokenSection1End
//
// A section can be given a custom id, as in the following example:
//
//	# Some Heading {#custom-id}
//
// - TokenSection1Begin "##"
// - TokenText "Some Heading"
// - TokenAttributeListBegin
// - TokenAttributeListID "custom-id"
// - TokenAttributeListEnd
// - TokenSection1Content
// - <...>
// - TokenSection1End
func (lx *Lexer) LexSection1() {
	Assert(lx.Peek1() == '#', "lexer state confused")
	hashes := lx.NextValids(CharInAny("#"))
	if len(hashes) != 1 {
		lx.Error(errors.New("expected section level 1"))
	}
	lx.Emit(TokenSection1Begin)
	lx.ExpectAndSkip(" ")
	lx.SkipWhitespaceNoNewLine()
	lx.LexTextUntilSpec(CharInAny("{\n")) // @todo: do we really want to allow all of these text elements inside a title? (wasn't there the same problem for sidenotes?)
	lx.SkipWhitespaceNoNewLine()
	if lx.Peek1() == '{' {
		lx.LexAttributeList()
	}
	lx.Emit(TokenSection1Content)
	lx.LexSection1Content()
	lx.Emit(TokenSection1End)
}

// LexSection1Content lexes the contents of a top level section.
// A section 1 can contain section twos, paragraphs, html elements, horizontal rule, {term,sidenote,link} definitions,
// code blocks, images, block quotes.
// A section 1 ends before another section 1 starts.
func (lx *Lexer) LexSection1Content() {
	lx.SkipWhitespace()
	for !lx.IsEOF() {
		if lx.IsTermDefinition() {
			lx.LexDefinitionList()
		} else if lx.IsHorizontalRule() {
			lx.LexHorizontalRule()
		} else {
			lx.SkipWhitespace()
			if lx.IsEOF() {
				break
			}
			if lx.IsTermDefinition() { // @todo: i don't like this
				lx.LexDefinitionList()
			} else if lx.IsHorizontalRule() {
				lx.LexHorizontalRule()
			} else if lx.Peek(3) == "```" {
				lx.LexCodeBlock()
			} else if lx.Peek(2) == "##" {
				lx.LexSection2()
			} else if lx.Peek(2) == "![" {
				lx.LexImage()
			} else if lx.Peek1() == '>' {
				lx.LexBlockQuotes()
			} else if lx.Peek1() == '#' {
				return // this section ends, next section starts
			} else if lx.Peek1() == '<' {
				lx.LexHtmlElement()
			} else if lx.Peek1() == '[' {
				lx.LexLinkOrSidenoteDefinition()
			} else {
				lx.LexParagraph()
			}
		}
	}
}

// LexSection2 lexes a second level section.
//
//	## Section 2
//
// - TokenSection2Begin "##"
// - TokenText "Section 2"
// - TokenSection2Content
// - <...>
// - TokenSection2End
//
// A section can be given a custom id, as in the following example:
//
//	## Some Heading {#custom-id}
//
// - TokenSection2Begin "##"
// - TokenText "Some Heading"
// - TokenAttributeListBegin
// - TokenAttributeListID "custom-id"
// - TokenAttributeListEnd
// - TokenSection2Content
// - <...>
// - TokenSection2End
func (lx *Lexer) LexSection2() {
	Assert(lx.Peek(2) == "##", "lexer state confused")
	hashes := lx.NextValids(CharInAny("#"))
	if len(hashes) != 2 {
		lx.Error(errors.New("expected section level 2"))
	}
	lx.Emit(TokenSection2Begin)
	lx.ExpectAndSkip(" ")
	lx.SkipWhitespaceNoNewLine()
	lx.LexTextUntilSpec(CharInAny("{\n")) // @todo: do we really want to allow all of these text elements inside a title? (wasn't there the same problem for sidenotes?)
	lx.SkipWhitespaceNoNewLine()
	if lx.Peek1() == '{' {
		lx.LexAttributeList()
	}
	lx.Emit(TokenSection2Content)
	lx.LexSection2Content()
	lx.Emit(TokenSection2End)
}

// LexSection2Content lees the contents of a second level section.
// A section 2 can contain paragraphs, html elements, horizontal rule, {term,sidenote,link} definitions, code blocks,
// images, block quotes.
// A section 2 ends before another section 1 or 2 starts.
func (lx *Lexer) LexSection2Content() {
	lx.SkipWhitespace()
	for !lx.IsEOF() {
		if lx.IsTermDefinition() {
			lx.LexDefinitionList()
		} else if lx.IsHorizontalRule() {
			lx.LexHorizontalRule()
		} else {
			lx.SkipWhitespace()
			if lx.IsEOF() {
				break
			}
			if lx.IsTermDefinition() { // @todo: i don't like this
				lx.LexDefinitionList()
			} else if lx.IsHorizontalRule() {
				lx.LexHorizontalRule()
			} else if lx.Peek(3) == "```" {
				lx.LexCodeBlock()
			} else if lx.Peek(2) == "##" {
				return // this section ends, next section starts
			} else if lx.Peek(2) == "![" {
				lx.LexImage()
			} else if lx.Peek1() == '>' {
				lx.LexBlockQuotes()
			} else if lx.Peek1() == '#' {
				return // this section ends, next section starts
			} else if lx.Peek1() == '<' {
				lx.LexHtmlElement()
			} else if lx.Peek1() == '[' {
				lx.LexLinkOrSidenoteDefinition()
			} else {
				lx.LexParagraph()
			}
		}
	}
}

func (lx *Lexer) LexParagraph() {
	lx.SkipWhitespace()
	lx.Emit(TokenParagraphBegin)
	lx.LexText()
	lx.Emit(TokenParagraphEnd)
}

// LexCodeBlock lexes a code block.
//
//	```
//	console.log('Hello, 世界')
//	```
//
// A code block can have an option language:
//
//	```js
//	console.log('Hello, 世界')
//	```
//
// A code block can also have an attribute list after the opening tag and optional language:
//
//	```js {link=https://gist.github.com/no/where lines=L1-2}
//	console.log('Hello, 世界')
//	console.log('Hello, 金星')
//	```
//
// The lexer produces the tokens:
// - TokenCodeBlockBegin "```"
// - TokenCodeBlockLang "go"
// - <tokens for attribute list if present>
// - TokenText "console.log('Hello, 世界')" // a single TokenText containing the full body of the code block
// - TokenCodeBlockEnd "```"
func (lx *Lexer) LexCodeBlock() {
	Assert(lx.Peek(3) == "```", "lexer state confused")
	lx.Next(3)
	lx.Emit(TokenCodeBlockBegin)
	lang := lx.NextValids(SpecAscii)
	if len(lang) > 0 {
		lx.Emit(TokenCodeBlockLang)
	}
	lx.SkipWhitespaceNoNewLine()
	if lx.Peek1() == '{' {
		lx.LexAttributeList()
	}
	lx.ExpectAndSkip("\n")
	for lx.Peek(3) != "```" {
		lx.NextUntilMatch("\n")
		//lx.Next1() // @todo: include newline or nah?
		lx.Emit(TokenText)
		lx.ExpectAndSkip("\n")
	}
	lx.Expect("```")
	lx.Emit(TokenCodeBlockEnd)
	lx.ExpectAndSkip("\n")
}

// LexAttributeList lexes an attribute list of the form
//
//	{key1=val1 key2 =val2 key3 = val3 key4 = 'with spaces between' key5= key6=val6}
//
// with a variable amount of key-value pairs.
// An empty attribute list looks like this: {}.
//
// The lexer produces the tokens:
// - TokenAttributeListBegin
// - TokenAttributeListKey "key1"
// - TokenText "val1"
// ...
// - TokenAttributeListKey "key4"
// - TokenText "with spaces between"
// - TokenAttributeListKey "key5"
// - TokenAttributeListKey "key6"
// - TokenText "val6"
// - TokenAttributeListEnd
//
// An attribute list can contain a custom id as the very first element:
//
//	{#some-id key1=val1 key2=val2}
//
// - TokenAttributeListBegin
// - TokenAttributeListID "some-id"
// - TokenAttributeListKey "key1"
// - TokenText "val1"
// - TokenAttributeListKey "key2"
// - TokenText "val2"
// - TokenAttributeListEnd
func (lx *Lexer) LexAttributeList() {
	Assert(lx.Peek1() == '{', "lexer state confused")
	lx.Next1()
	lx.Emit(TokenAttributeListBegin)
	lx.SkipWhitespace()
	if lx.Peek1() == '#' {
		lx.SkipNext1()
		id := lx.NextValids(SpecAttrVal)
		if len(id) == 0 {
			lx.Error(errors.New("must provide id"))
		}
		lx.Emit(TokenAttributeListID)
		lx.SkipWhitespace()
	}
	for !lx.IsEOF() && lx.Peek1() != '}' {
		key := lx.NextValids(SpecAttrKey)
		if len(key) == 0 {
			lx.Error(errors.New("must provide a key"))
		}
		lx.Emit(TokenAttributeListKey)
		lx.SkipWhitespaceNoNewLine()
		if lx.MatchAtPos("=") {
			lx.SkipNext1()
			lx.SkipWhitespaceNoNewLine()
			if lx.Peek1() != '}' {
				lx.LexStringValue()
			}
		}
		lx.SkipWhitespace()
	}
	lx.Expect("}")
	lx.Emit(TokenAttributeListEnd)
}

// LexStringValue lexes a string value that is either a single word like
//
//	oneword
//
// or one or multiple words enclosed in single or double quotes
//
//	'one or multiple words enclosed in single quotes'
//
// or simply empty (the absence of any characters or ”).
//
// The lexer produces the token:
// - TokenText "one or multiple words enclosed in single quotes"
//
// No token is produced if the string value is empty.
func (lx *Lexer) LexStringValue() { // @todo: rename this function, since it is specific to LexAttributeList
	if lx.Peek1() == '\'' {
		lx.Next1()
		lx.Skip()
		lx.NextUntilMatch("'")
		lx.EmitIfNonEmpty(TokenText)
		lx.ExpectAndSkip("'")
	} else if lx.Peek1() == '"' {
		lx.Next1()
		lx.Skip()
		lx.NextUntilMatch("\"")
		lx.EmitIfNonEmpty(TokenText)
		lx.ExpectAndSkip("\"")
	} else {
		lx.NextValids(SpecAttrVal)
		lx.EmitIfNonEmpty(TokenText)
	}
}

// LexSingleWordSidenote lexes sidenotes that wrap around only a single word.
//
// If the sidenote is only around a single word, it can be written like this:
//
//	Link-Text[^0]
//
// - TokenLinkableBegin
// - TokenText "Link-Text"
// - TokenSidenoteRef "0"
// - TokenLinkableEnd
//
// or
//
//	Link-Text(^Sidenote content.)
//
// - TokenLinkableBegin
// - TokenText "Link-Text"
// - TokenSidenoteContent
// - TokenText "Sidenote content."
// - TokenLinkableEnd
func (lx *Lexer) LexSingleWordSidenote() {
	Assert(SpecNonWhitespace.IsValid(lx.Peek1()), "lexer state confused")
	lx.Emit(TokenLinkableBegin)
	lx.NextValids(SpecSingleWordSidenote)
	lx.Emit(TokenText)
	if lx.Peek(2) == "[^" {
		lx.SkipNext(2)
		lx.NextUntilMatch("]")
		lx.Emit(TokenSidenoteRef)
		lx.ExpectAndSkip("]")
	} else if lx.Peek(2) == "(^" {
		lx.SkipNext(2)
		lx.Emit(TokenSidenoteContent)
		lx.LexTextUntil(")") // @todo: LexSidenoteText or sth, Sidenotes shouldn't allow all text types?
		lx.ExpectAndSkip(")")
	} else {
		lx.Error(errors.New("expected [^ or (^"))
	}
	lx.Emit(TokenLinkableEnd)
}

// LexLinkOrSidenote lexes the following elements:
//
//	[Link Text](https://example.com)
//
// - TokenLinkableBegin "["
// - TokenText "Link Text"
// - TokenLinkHref "https://example.com"
// - TokenLinkableEnd
//
//	[Link Text][0]
//
// - TokenLinkableBegin "["
// - TokenText "Link Text"
// - TokenLinkRef "0"
// - TokenLinkableEnd
//
//	[Sidenote Text][^0]
//
// - TokenLinkableBegin "["
// - TokenText "Sidenote Text"
// - TokenSidenoteRef "0"
// - TokenLinkableEnd
//
// The following is allowed, it makes an <a> with an empty href
//
//	[Link Text]
//
// - TokenLinkableBegin "["
// - TokenText "Link Text"
// - TokenLinkHref ""
// - TokenLinkableEnd
//
// A bit exotic, but also allowed
//
//	[Sidenote Text](^Sidenote content)
//
// - TokenLinkableBegin "["
// - TokenText "Sidenote Text"
// - TokenSidenoteContent
// - TokenText "Sidenote content"
// - TokenLinkableEnd
func (lx *Lexer) LexLinkOrSidenote() {
	Assert(lx.Peek1() == '[', "lexer state confused")
	lx.Next1()
	lx.Emit(TokenLinkableBegin)
	lx.LexTextUntil("]")
	lx.Expect("]")
	if lx.Peek(2) == "[^" {
		// reference style sidenote
		lx.SkipNext(2)
		lx.NextUntilMatch("]")
		lx.Emit(TokenSidenoteRef)
		lx.ExpectAndSkip("]")
	} else if lx.Peek1() == '[' {
		// reference style link
		lx.SkipNext1()
		lx.NextUntilMatch("]")
		lx.Emit(TokenLinkRef)
		lx.ExpectAndSkip("]")
	} else if lx.Peek(2) == "(^" { // @todo: good idea to make my own MD extension?
		// (normal) sidenote
		lx.SkipNext(2)
		lx.Emit(TokenSidenoteContent)
		lx.LexTextUntil(")") // @todo: LexSidenoteText or sth, Sidenotes shouldn't allow all text types?
		lx.ExpectAndSkip(")")
	} else if lx.Peek1() == '(' {
		// (normal) link
		lx.SkipNext1()
		lx.NextUntilMatch(")")
		lx.Emit(TokenLinkHref)
		lx.ExpectAndSkip(")")
	} else {
		// not an error, it's just an link with an empty href
		lx.Skip()
		lx.Emit(TokenLinkHref)
	}
	lx.Skip() // just for good measure
	lx.Emit(TokenLinkableEnd)
}

// LexText lexes text elements, namely:
// - strings
// - <https://example.com/> form links
// - *emphasis*, **strong**, ***emphasis strong***
// - _emphasis_, __strong__, ___emphasis strong___
// - &...; specials and ~ (nbsp), -- (en dash), --- (em dash), - (hyphen)
// - ~~strikethrough~~ and ~strikethrough~
// - `mono spaced text`
// - "enquote" and 'enquote' and <<enquote>>
// - take care of escaped syntax \<>
//
// LexText stops lexing once the string at Peek is a syntactic construct that
// denotes a content (non-text) element.
func (lx *Lexer) LexText() {
	for !lx.IsEOF() && !lx.IsParagraphEnder() {
		if ok, _ := lx.IsAmpSpecial(); ok {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexAmpSpecial()
		} else if lx.IsSingleWordSidenote() {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexSingleWordSidenote()
		} else if lx.Peek(3) == "***" || lx.Peek(3) == "___" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEmphasisStrong()
		} else if lx.Peek(2) == "**" || lx.Peek(2) == "__" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexStrong()
		} else if lx.Peek(2) == "<<" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEnquoteAngled()
		} else if lx.Peek(2) == "~~" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexStrikethrough()
		} else if lx.Peek(2) == "==" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexMarker()
		} else if lx.Peek1() == '*' || lx.Peek1() == '_' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEmphasis()
		} else if lx.Peek1() == '<' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexLinkifyOrHtmlElement()
		} else if lx.Peek1() == '`' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexMono()
		} else if lx.Peek1() == '"' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEnquoteDouble()
			//} else if lx.Peek1() == '`' {
			//	lx.EmitIfNonEmpty(TokenText)
			//	lx.LexEnquoteSingle()
		} else if lx.Peek1() == '[' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexLinkOrSidenote()
		} else if lx.IsEscape() {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEscape()
		} else {
			lx.Next1()
		}
	}
	lx.EmitIfNonEmpty(TokenText)
}

func (lx *Lexer) LexEscape() {
	Assert(lx.Peek1() == '\\' && lx.IsEscape(), "lexer state confused")
	lx.SkipNext1()
	switch lx.Peek1() {
	case '\\', '!', '`', '*', '_', '{', '}', '<', '>', '[', ']', '(', ')', '|', '#', '+', '-', '.', '&':
		lx.Next1()
		lx.Emit(TokenText)
	default:
		panic("unreachable (programmer messed up)")
	}
}

// LexTextUntil lexes text elements, namely:
// - strings
// - <https://example.com/> form links
// - *emphasis*, **strong**, ***emphasis strong***
// - _emphasis_, __strong__, ___emphasis strong___
// - &...; specials and ~ (nbsp), -- (en dash), --- (em dash), - (hyphen)
// - ~~strikethrough~~ and ~strikethrough~
// - `mono spaced text`
// - "enquote" and 'enquote' and <<enquote>>
// - take care of escaped syntax \<>
//
// LexTextUntil stops lexing once match matches at Peek.
func (lx *Lexer) LexTextUntil(match string) {
	for !lx.IsEOF() && !lx.MatchAtPos(match) {
		if ok, _ := lx.IsAmpSpecial(); ok {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexAmpSpecial()
		} else if lx.Peek(3) == "***" || lx.Peek(3) == "___" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEmphasisStrong()
		} else if lx.Peek(2) == "**" || lx.Peek(2) == "__" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexStrong()
		} else if lx.Peek(2) == "<<" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEnquoteAngled()
		} else if lx.Peek(2) == "~~" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexStrikethrough()
		} else if lx.Peek(2) == "==" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexMarker()
		} else if lx.Peek1() == '*' || lx.Peek1() == '_' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEmphasis()
		} else if lx.Peek1() == '<' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexLinkify()
		} else if lx.Peek1() == '`' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexMono()
		} else if lx.Peek1() == '"' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEnquoteDouble()
			//} else if lx.Peek1() == '`' {
			//	lx.EmitIfNonEmpty(TokenText)
			//	lx.LexEnquoteSingle()
		} else if lx.Peek1() == '\\' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEscape()
		} else {
			lx.Next1()
		}
	}
	lx.EmitIfNonEmpty(TokenText)
}

// LexTextUntilSpec lexes text elements, namely:
// - strings
// - <https://example.com/> form links
// - *emphasis*, **strong**, ***emphasis strong***
// - _emphasis_, __strong__, ___emphasis strong___
// - &...; specials and ~ (nbsp), -- (en dash), --- (em dash), - (hyphen)
// - ~~strikethrough~~ and ~strikethrough~
// - `mono spaced text`
// - "enquote" and 'enquote' and <<enquote>>
// - take care of escaped syntax \<>
//
// LexTextUntilSpec stops lexing once the spec matches at Peek.
func (lx *Lexer) LexTextUntilSpec(spec CharSpec) {
	for !lx.IsEOF() && !spec.IsValid(lx.Peek1()) {
		if ok, _ := lx.IsAmpSpecial(); ok {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexAmpSpecial()
		} else if lx.Peek(3) == "***" || lx.Peek(3) == "___" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEmphasisStrong()
		} else if lx.Peek(2) == "**" || lx.Peek(2) == "__" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexStrong()
		} else if lx.Peek(2) == "<<" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEnquoteAngled()
		} else if lx.Peek(2) == "~~" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexStrikethrough()
		} else if lx.Peek(2) == "==" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexMarker()
		} else if lx.Peek1() == '*' || lx.Peek1() == '_' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEmphasis()
		} else if lx.Peek1() == '<' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexLinkify()
		} else if lx.Peek1() == '`' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexMono()
		} else if lx.Peek1() == '"' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEnquoteDouble()
			//} else if lx.Peek1() == '`' {
			//	lx.EmitIfNonEmpty(TokenText)
			//	lx.LexEnquoteSingle()
		} else if lx.Peek1() == '\\' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEscape()
		} else {
			lx.Next1()
		}
	}
	lx.EmitIfNonEmpty(TokenText)
}

// LexTextUntilPred lexes text elements, namely:
// - strings
// - <https://example.com/> form links
// - *emphasis*, **strong**, ***emphasis strong***
// - _emphasis_, __strong__, ___emphasis strong___
// - &...; specials and ~ (nbsp), -- (en dash), --- (em dash), - (hyphen)
// - ~~strikethrough~~ and ~strikethrough~
// - `mono spaced text`
// - "enquote" and 'enquote' and <<enquote>>
// - take care of escaped syntax \<>
//
// LexTextUntilPred stops lexing when the predicate returns true.
func (lx *Lexer) LexTextUntilPred(pred Predicate) {
	for !lx.IsEOF() && !pred() {
		if ok, _ := lx.IsAmpSpecial(); ok {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexAmpSpecial()
		} else if lx.Peek(3) == "***" || lx.Peek(3) == "___" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEmphasisStrong()
		} else if lx.Peek(2) == "**" || lx.Peek(2) == "__" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexStrong()
		} else if lx.Peek(2) == "<<" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEnquoteAngled()
		} else if lx.Peek(2) == "~~" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexStrikethrough()
		} else if lx.Peek(2) == "==" {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexMarker()
		} else if lx.Peek1() == '*' || lx.Peek1() == '_' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEmphasis()
		} else if lx.Peek1() == '<' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexLinkify()
		} else if lx.Peek1() == '`' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexMono()
		} else if lx.Peek1() == '"' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEnquoteDouble()
			//} else if lx.Peek1() == '`' {
			//	lx.EmitIfNonEmpty(TokenText)
			//	lx.LexEnquoteSingle()
		} else if lx.Peek1() == '\\' {
			lx.EmitIfNonEmpty(TokenText)
			lx.LexEscape()
		} else {
			lx.Next1()
		}
	}
	lx.EmitIfNonEmpty(TokenText)
}

func (lx *Lexer) LexLinkifyOrHtmlElement() {
	Assert(lx.Peek1() == '<', "lexer state confused")
	if lx.MatchAtPos("<http://") || lx.MatchAtPos("<https://") || lx.MatchAtPos("<mailto:") { // @todo: what about other protocols?
		lx.LexLinkify()
	} else {
		lx.LexHtmlElement()
	}
}

// LexLinkify lexes linkified text like <https://example.com/some/path>
//
// - TokenLinkify "https://example.com/some/path"
func (lx *Lexer) LexLinkify() {
	Assert(lx.Peek1() == '<', "lexer state confused")
	lx.SkipNext1()
	lx.NextUntilMatch(">")
	lx.Emit(TokenLinkify)
	lx.ExpectAndSkip(">")
}

// LexLinkOrSidenoteDefinition lexes the definition of the link or sidenote.
//
//	[0]: https://example.com
//
// - TokenLinkDef "0"
// - TokenText "https://example.com"
//
//	[^0]: This is the side note content.
//
// - TokenSidenoteDef "0"
// - TokenText "This is the sidenote content."
// - TokenSidenoteDefEnd
func (lx *Lexer) LexLinkOrSidenoteDefinition() {
	Assert(lx.Peek1() == '[', "lexer state confused")
	lx.SkipNext1()
	if lx.Peek1() == '^' {
		// sidenote definition
		lx.SkipNext1()
		lx.NextUntilMatch("]")
		lx.Emit(TokenSidenoteDef)
		lx.ExpectAndSkip("]:")
		lx.SkipWhitespaceNoNewLine()
		lx.LexTextUntil("\n")
		lx.ExpectAndSkip("\n")
		lx.Emit(TokenSidenoteDefEnd)
	} else {
		// link definition
		lx.NextUntilMatch("]")
		lx.Emit(TokenLinkDef)
		lx.ExpectAndSkip("]:")
		lx.SkipWhitespaceNoNewLine()
		lx.NextUntilMatch("\n")
		lx.Emit(TokenText)
		lx.ExpectAndSkip("\n")
	}
}

// LexImage lexes an image of the form
//
//	![Alt Text](/path/to/image)
//
// - TokenImageBegin "!["
// - TokenImageAltText "Alt Text"
// - TokenImagePath "/path/to/image"
// - TokenImageEnd
//
//	![Alt Text](/path/to/image "Optional Image Title")
//
// - TokenImageBegin "!["
// - TokenImageAltText "Alt Text"
// - TokenImagePath "/path/to/image"
// - TokenImageTitle "Optional Image Title"
// - TokenImageEnd
func (lx *Lexer) LexImage() {
	Assert(lx.Peek(2) == "![", "lexer state confused")
	lx.Next(2)
	lx.Emit(TokenImageBegin)
	lx.NextUntilMatch("]")
	lx.Emit(TokenImageAltText)
	lx.ExpectAndSkip("]")
	if lx.Expect("(") {
		lx.SkipWhitespaceNoNewLine()
		if lx.Peek1() == '\'' {
			lx.NextUntilMatch("'")
			lx.Emit(TokenImagePath)
			lx.ExpectAndSkip("'")
		} else if lx.Peek1() == '"' {
			lx.NextUntilMatch("\"")
			lx.Emit(TokenImagePath)
			lx.ExpectAndSkip("\"")
		} else {
			lx.NextValids(SpecImagePath)
			lx.Emit(TokenImagePath)
		}
		lx.SkipWhitespaceNoNewLine()
		if lx.Peek1() == '"' {
			lx.Next1()
			lx.Skip()
			lx.NextUntilMatch(`"`)
			lx.Emit(TokenImageTitle)
			lx.ExpectAndSkip(`"`)
		}
		lx.SkipWhitespaceNoNewLine()
		lx.ExpectAndSkip(`)`)
	}
	lx.Emit(TokenImageEnd)
}

// LexBlockQuotes lexes block quotes of the form
//
//	> Elea acta est.
//	> -- Author, Source
//
// - TokenBlockquoteBegin
// - TokenText "Elea acta est."
// - TokenBlockquoteAttrAuthor
// - TokenText "Author"
// - TokenBlockquoteAttrSource
// - TokenText "Source"
// - TokenBlockquoteAttrEnd
// - TokenBlockquoteEnd
//
// @todo: nested block quotes
func (lx *Lexer) LexBlockQuotes() {
	Assert(lx.Peek1() == '>', "lexer state confused")
	lx.Emit(TokenBlockquoteBegin)
	for !lx.IsEOF() && lx.Peek1() == '>' {
		lx.SkipNext1()
		lx.SkipWhitespaceNoNewLine()
		if lx.Peek(2) == "--" {
			lx.SkipNext(2)
			lx.Emit(TokenBlockquoteAttrAuthor)
			lx.SkipWhitespaceNoNewLine()
			lx.LexTextUntilSpec(CharInAny(",\n"))
			if lx.Peek1() == ',' {
				lx.SkipNext1()
				lx.SkipWhitespaceNoNewLine()
				lx.Emit(TokenBlockquoteAttrSource)
				lx.LexTextUntil("\n")
			}
			lx.Emit(TokenBlockquoteAttrEnd)
			lx.ExpectAndSkip("\n")
			lx.Emit(TokenBlockquoteEnd)
			lx.SkipWhitespaceNoNewLine()
			if lx.Peek1() == '>' {
				lx.Error(errors.New("blockquote already finished by attribution"))
			}
			return
		}
		lx.LexTextUntil("\n")
		lx.Expect("\n")
		lx.SkipWhitespaceNoNewLine()
	}
	lx.Emit(TokenBlockquoteEnd)
}

// LexHorizontalRule lexes a horizontal rule of the form <WSL>---<WSL> or <WSL>***<WSL>
// where <WSL> denotes Whitespace including at least one newline.
//
// - TokenHorizontalRule "---" or "***"
func (lx *Lexer) LexHorizontalRule() {
	lx.SkipWhitespaceNoNewLine()
	lx.Expect("\n")
	lx.SkipWhitespace()
	if !(lx.NextIfMatch("---") || lx.NextIfMatch("***")) {
		lx.Error(errors.New("expected horizontal rule `---` or `***`"))
	}
	lx.Emit(TokenHorizontalRule)
	lx.SkipWhitespace()
}

// LexDefinitionList lexes a definition of the form
//
// Term
// : Explanation of term
//
// - TokenDefinitionTerm "Term"
// - TokenDefinitionExplanationBegin
// - TokenText "Explanation of term"
// - TokenDefinitionExplanationEnd
// - <...> optional more term definitions
func (lx *Lexer) LexDefinitionList() {
	Assert(lx.IsStartOfLine(), "lexer state confused")
	for lx.IsTermDefinition() {
		term := lx.NextValids(SpecNonWhitespace)
		if len(term) == 0 {
			lx.Error(errors.New("term missing from definition"))
		}
		lx.Emit(TokenDefinitionTerm)
		lx.SkipWhitespaceNoNewLine()
		lx.Expect("\n")
		lx.SkipWhitespace()
		lx.Expect(":")
		lx.SkipWhitespace()
		lx.Emit(TokenDefinitionExplanationBegin)
		lx.LexTextUntil("\n")
		lx.ExpectAndSkip("\n")
		lx.Emit(TokenDefinitionExplanationEnd)
	}
}

// LexMono lexes a monospaced text block like
//
//	`This text is in monospace`
//
// - TokenMono "This text is in monospace"
func (lx *Lexer) LexMono() {
	Assert(lx.Peek1() == '`', "lexer state confused")
	lx.SkipNext1()
	lx.NextUntilMatch("`")
	lx.Emit(TokenMono)
	lx.ExpectAndSkip("`")
}

func (lx *Lexer) LexEmphasis() {
	Assert(lx.Peek1() == '*' || lx.Peek1() == '_', "lexer state confused")
	lx.Next1()
	lx.Emit(TokenEmphasisBegin)
	lx.LexTextUntilPred(lx.IsEmphasis)
	lx.Next(1) // should do two different functions for * and _
	lx.Emit(TokenEmphasisEnd)
}

func (lx *Lexer) LexStrong() {
	Assert(lx.Peek(2) == "**" || lx.Peek(2) == "__", "lexer state confused")
	lx.Next(2)
	lx.Emit(TokenStrongBegin)
	lx.LexTextUntilPred(lx.IsStrong)
	lx.Next(2) // should do two different functions for * and _
	lx.Emit(TokenStrongEnd)
}

func (lx *Lexer) LexEmphasisStrong() {
	Assert(lx.Peek(3) == "***" || lx.Peek(3) == "___", "lexer state confused")
	lx.Next(3)
	lx.Emit(TokenEmphasisStrongBegin)
	lx.LexTextUntilPred(lx.IsEmphasisStrong)
	lx.Next(3) // should do two different functions for * and _
	lx.Emit(TokenEmphasisStrongEnd)
}

func (lx *Lexer) LexEnquoteSingle() {
	Assert(lx.Peek1() == '`', "lexer state confused")
	lx.Next1()
	lx.Emit(TokenEnquoteSingleBegin)
	lx.LexTextUntil("'")
	lx.Expect("'")
	lx.Emit(TokenEnquoteSingleEnd)
}

func (lx *Lexer) LexEnquoteDouble() {
	Assert(lx.Peek1() == '"', "lexer state confused")
	lx.Next1()
	lx.Emit(TokenEnquoteDoubleBegin)
	lx.LexTextUntil("\"")
	lx.Expect("\"")
	lx.Emit(TokenEnquoteDoubleEnd)
}

func (lx *Lexer) LexEnquoteAngled() {
	Assert(lx.Peek(2) == "<<", "lexer state confused")
	lx.Next(2)
	lx.Emit(TokenEnquoteAngledBegin)
	lx.LexTextUntil(">>")
	lx.Expect(">>")
	lx.Emit(TokenEnquoteAngledEnd)
}

func (lx *Lexer) LexStrikethrough() {
	Assert(lx.Peek(2) == "~~", "lexer state confused")
	lx.Next(2)
	lx.Emit(TokenStrikethroughBegin)
	lx.LexTextUntil("~~")
	lx.Expect("~~")
	lx.Emit(TokenStrikethroughEnd)
}

func (lx *Lexer) LexMarker() {
	Assert(lx.Peek(2) == "==", "lexer state confused")
	lx.Next(2)
	lx.Emit(TokenMarkerBegin)
	lx.LexTextUntil("==")
	lx.Expect("==")
	lx.Emit(TokenMarkerEnd)
}

// LexHtmlElement lexes an HTML element like
//
//	<tag-name attr="val" ...>
//	   ...
//	</tag-name>
//
// or
//
//	<tag-name attr="val" ...>
//
// an attribute does not need to have a value
//
//	<tag-name attr>
func (lx *Lexer) LexHtmlElement() {
	Assert(lx.Peek1() == '<', "lexer state confused")
	lx.SkipNext1()
	tag := lx.NextValids(CharInSpec{SpecAscii, CharInAny("-_")})
	if len(tag) == 0 {
		lx.Error(errors.New("expected html tag name"))
	}
	lx.Emit(TokenHtmlTagOpen)
	lx.SkipWhitespace()
	if lx.MatchAtPos(">") {
		lx.SkipNext1()
	} else {
		lx.LexHtmlElementAttributes()
		lx.ExpectAndSkip(">")
	}
	lx.LexHtmlElementContent(tag)
}

func (lx *Lexer) LexHtmlElementAttributes() {
	lx.SkipWhitespace()
	for lx.Peek1() != '>' {
		attrKey := lx.NextValids(CharInSpec{SpecAscii, CharInAny("-_")})
		if len(attrKey) == 0 {
			lx.Error(fmt.Errorf("expected attribute or >, got: %s", lx.Peek(7)))
			break
		}
		lx.Emit(TokenHtmlTagAttrKey)
		lx.SkipWhitespace()
		if lx.Peek1() == '=' {
			lx.SkipNext1()
			lx.ExpectAndSkip("\"")
			attrVal, _ := lx.NextUntilMatch("\"")
			_ = attrVal
			lx.Emit(TokenHtmlTagAttrVal)
			lx.ExpectAndSkip("\"")
		}
		lx.SkipWhitespace()
	}
}

func (lx *Lexer) LexHtmlElementContent(tag string) {
	lx.Emit(TokenHtmlTagContent)
	if lx.HasCloseTag(tag) {
		for !lx.IsEOF() {
			if lx.Peek(2) == "</" {
				lx.SkipNext(2)
				lx.SkipWhitespace()
				mtag := lx.NextValids(CharInSpec{SpecAscii, CharInAny("-_")})
				if tag != mtag {
					lx.Error(fmt.Errorf("expected <%s> to be closed first, got: </%s>", tag, mtag))
				}
				lx.SkipWhitespace()
				lx.ExpectAndSkip(">")
				break
			}
			// @todo: first check if there is a nested html element
			lx.LexParagraph()
		}
	}
	lx.Emit(TokenHtmlTagClose)
}

func (lx *Lexer) HasCloseTag(tag string) bool {
	// @todo: this is so incredibly stupid
	// @todo: this also doesn't work with nested elements
	lpos := lx.Pos
	lcon := lx.Consumed
	defer func() {
		lx.Pos = lpos
		lx.Consumed = lcon
	}()
	for !lx.IsEOF() {
		if lx.MatchAtPos(`\</`) {
			lx.Next(3)
		} else if lx.MatchAtPos("</") {
			lx.SkipNext(2)
			lx.SkipWhitespace()
			if lx.MatchAtPos(tag) {
				lx.SkipNext(len(tag))
				lx.SkipWhitespace()
				return lx.MatchAtPos(">")
			}
		} else {
			lx.Next1()
		}
	}
	return false
}

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
