package lexer_test

import (
	"testing"

	"github.com/go-test/deep"

	//"github.com/cvanloo/blog-go/markup"
	"github.com/cvanloo/blog-go/markup/lexer"
)

/*
func TestLexer(t *testing.T) {
	testBlog := markup.BlogTestSource
	testExpected := markup.LexerTestTokens
	lx := lexer.New()
	if err := lx.LexSource("testBlog", testBlog); err != nil {
		for _, tok := range lx.Lexemes {
			t.Log(tok)
		}
		t.Errorf("got %d errors, first error is: %s", len(lx.Errors), err)
		for i, err := range lx.Errors {
			t.Errorf("error %d: %s", i, err)
		}
		t.FailNow()
	}
	if len(lx.Lexemes) != len(testExpected) {
		t.Errorf("expected %d tokens, got %d tokens", len(lx.Lexemes), len(testExpected))
		l := min(len(lx.Lexemes), len(testExpected))
		for i := range l {
			if lx.Lexemes[i].Type != testExpected[i].Type {
				t.Errorf("tokens don't match at index: %d, expected: %s, got: %s", i, testExpected[i].Type, lx.Lexemes[i].Type)
			}
		}
		if len(lx.Lexemes) > len(testExpected) {
			c := len(lx.Lexemes) - len(testExpected)
			t.Errorf("lexer produced too many tokens: +%d", c)
			for i := range c {
				t.Errorf("unexpected token: %s", lx.Lexemes[len(lx.Lexemes)-1+i].Type)
			}
		} else {
			c := len(testExpected) - len(lx.Lexemes)
			t.Errorf("lexer produced too few tokens: -%d", c)
			for i := range c {
				t.Errorf("missing token: %s", testExpected[len(testExpected)-1+i].Type)
			}
		}
	}
	for i, tok := range lx.Lexemes {
		e := testExpected[i]
		if e.Type != tok.Type {
			t.Errorf("wrong token type at index: %d, expected: %s, got: %s", i, e.Type, tok.Type)
		}
		if e.Text != "" && tok.Text != e.Text {
			t.Errorf("error at index: %d, token: %s, expected: `%s`, got: `%s`", i, tok.Type, e.Text, tok.Text)
		}
	}
}*/

func TestLexMeta(t *testing.T) {
	var testCases = []TestCase{
		{
			name: "Empty Meta Block",
			source: `---
---
`,
			expected: []lexer.Token{
				{Type: lexer.TokenMetaBegin, Text: "---"},
				{Type: lexer.TokenMetaEnd, Text: "---"},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Empty Meta Block with Space Between",
			source: `---

---
`,
			expected: []lexer.Token{
				{Type: lexer.TokenMetaBegin, Text: "---"},
				{Type: lexer.TokenMetaEnd, Text: "---"},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Empty Meta Block with White Space Before",
			source: `

    ---
---
`,
			expected: []lexer.Token{
				{Type: lexer.TokenMetaBegin, Text: "---"},
				{Type: lexer.TokenMetaEnd, Text: "---"},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Empty Meta Block with White Space Around Markers",
			source: `

    --- 	 
 --- 	   
`,
			expected: []lexer.Token{
				{Type: lexer.TokenMetaBegin, Text: "---"},
				{Type: lexer.TokenMetaEnd, Text: "---"},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Empty Meta Block with White Space After",
			source: `---
---	 

    
			`,
			expected: []lexer.Token{
				{Type: lexer.TokenMetaBegin, Text: "---"},
				{Type: lexer.TokenMetaEnd, Text: "---"},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Empty Meta Block Without End Marker",
			source: `---
`,
			expected: []lexer.Token{
				{Type: lexer.TokenMetaBegin, Text: "---"},
				{Type: lexer.TokenMetaEnd, Text: ""}, // still emitted, as the lexer tries to recover
				{Type: lexer.TokenEOF, Text: ""},
			},
			expectedErrors: []string{
				"expected: `---`, got: ``",
				"expected: `\\n`, got: ``", // @todo: better errors
			},
		},
		{
			name: "Non-Empty Meta Block",
			source: `
---
foo: bar baz
---
`,
			expected: []lexer.Token{
				{Type: lexer.TokenMetaBegin, Text: "---"},
				{Type: lexer.TokenMetaKey, Text: "foo"},
				{Type: lexer.TokenText, Text: "bar baz"},
				{Type: lexer.TokenMetaEnd, Text: "---"},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Non-Empty Meta Block with Keys, Missing End Marker",
			source: `
---
foo: bar baz
oof: rab zab

`,
			expected: []lexer.Token{
				{Type: lexer.TokenMetaBegin, Text: "---"},
				{Type: lexer.TokenMetaKey, Text: "foo"},
				{Type: lexer.TokenText, Text: "bar baz"},
				{Type: lexer.TokenMetaKey, Text: "oof"},
				{Type: lexer.TokenText, Text: "rab zab"},
				{Type: lexer.TokenMetaEnd, Text: ""}, // still emitted, as the lexer tries to recover
				{Type: lexer.TokenEOF, Text: ""},
			},
			expectedErrors: []string{
				"expected: `---`, got: ``",
				"expected: `\\n`, got: ``", // @todo: better errors
			},
		},
		{
			name: "Non-Empty Meta Block using +++ instead of ---",
			source: `
+++
foo: bar baz
oof: rab zab
+++
`,
			expected: []lexer.Token{
				{Type: lexer.TokenMetaBegin, Text: "+++"},
				{Type: lexer.TokenMetaKey, Text: "foo"},
				{Type: lexer.TokenText, Text: "bar baz"},
				{Type: lexer.TokenMetaKey, Text: "oof"},
				{Type: lexer.TokenText, Text: "rab zab"},
				{Type: lexer.TokenMetaEnd, Text: "+++"},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Non-Empty Meta Block with mixed +++ and ---",
			source: `
+++
foo: bar baz
oof: rab zab
---
`,
			expected: []lexer.Token{
				{Type: lexer.TokenMetaBegin, Text: "+++"},
				{Type: lexer.TokenMetaKey, Text: "foo"},
				{Type: lexer.TokenText, Text: "bar baz"},
				{Type: lexer.TokenMetaKey, Text: "oof"},
				{Type: lexer.TokenText, Text: "rab zab"},
				{Type: lexer.TokenMetaEnd, Text: "---"},
				{Type: lexer.TokenEOF, Text: ""},
			},
			expectedErrors: []string{
				"expected: `+++`, got: `---`",
			},
		},
		{
			name: "Meta Block with +++, =, and Amp Specials",
			source: `
+++
author = Colin van~Loo
+++
`,
			expected: []lexer.Token{
				{Type: lexer.TokenMetaBegin, Text: "+++"},
				{Type: lexer.TokenMetaKey, Text: "author"},
				{Type: lexer.TokenText, Text: "Colin van"},
				{Type: lexer.TokenAmpSpecial, Text: "~"},
				{Type: lexer.TokenText, Text: "Loo"},
				{Type: lexer.TokenMetaEnd, Text: "+++"},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
	}
	RunTests(t, testCases)
}

func TestLexSectionHeader(t *testing.T) {
	testCases := []TestCase{
		{
			name:   "Section 1 Header",
			source: `# Hello, World!`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Hello, World!"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Section 2 Header",
			source: `
# こんにちは、世界！

## Hello, World!
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "こんにちは、世界！"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenSection2Begin, Text: "##"},
				{Type: lexer.TokenText, Text: "Hello, World!"},
				{Type: lexer.TokenSection2Content, Text: ""},
				{Type: lexer.TokenSection2End, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Section 2 Header With Whitespace Around",
			source: `   
#   Goodnight, Moon! 

##   	  Hello, World!   	
  `,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Goodnight, Moon! "},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenSection2Begin, Text: "##"},
				{Type: lexer.TokenText, Text: "Hello, World!   \t"},
				{Type: lexer.TokenSection2Content, Text: ""},
				{Type: lexer.TokenSection2End, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Section headers containing amp specials",
			source: `   
#   Goodnight,~Moon! 

##   	  &ldquo;Hello, --- World...   	
  `,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Goodnight,"},
				{Type: lexer.TokenAmpSpecial, Text: "~"},
				{Type: lexer.TokenText, Text: "Moon! "},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenSection2Begin, Text: "##"},
				{Type: lexer.TokenAmpSpecial, Text: "&ldquo;"},
				{Type: lexer.TokenText, Text: "Hello, "},
				{Type: lexer.TokenAmpSpecial, Text: "---"},
				{Type: lexer.TokenText, Text: " World"},
				{Type: lexer.TokenAmpSpecial, Text: "..."},
				{Type: lexer.TokenText, Text: "   \t"},
				{Type: lexer.TokenSection2Content, Text: ""},
				{Type: lexer.TokenSection2End, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Section headers with custom id",
			source: `
# Section 1 {#section-1}
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 1 "},
				{Type: lexer.TokenAttributeListBegin, Text: "{"},
				{Type: lexer.TokenAttributeListID, Text: "section-1"},
				{Type: lexer.TokenAttributeListEnd, Text: "}"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Section headers with attributes",
			source: `
# Section 1 {key1=val1 key2 key3='val 3' key4 = "val 4" key5 =}
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 1 "},
				{Type: lexer.TokenAttributeListBegin, Text: "{"},
				{Type: lexer.TokenAttributeListKey, Text: "key1"},
				{Type: lexer.TokenText, Text: "val1"},
				{Type: lexer.TokenAttributeListKey, Text: "key2"},
				{Type: lexer.TokenAttributeListKey, Text: "key3"},
				{Type: lexer.TokenText, Text: "val 3"},
				{Type: lexer.TokenAttributeListKey, Text: "key4"},
				{Type: lexer.TokenText, Text: "val 4"},
				{Type: lexer.TokenAttributeListKey, Text: "key5"},
				{Type: lexer.TokenAttributeListEnd, Text: "}"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Section headers with custom id and attributes",
			source: `
# Section 1 {#section-1 key1=val1 key2 key3='val 3' key4 = "val 4" key5 =}
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 1 "},
				{Type: lexer.TokenAttributeListBegin, Text: "{"},
				{Type: lexer.TokenAttributeListID, Text: "section-1"},
				{Type: lexer.TokenAttributeListKey, Text: "key1"},
				{Type: lexer.TokenText, Text: "val1"},
				{Type: lexer.TokenAttributeListKey, Text: "key2"},
				{Type: lexer.TokenAttributeListKey, Text: "key3"},
				{Type: lexer.TokenText, Text: "val 3"},
				{Type: lexer.TokenAttributeListKey, Text: "key4"},
				{Type: lexer.TokenText, Text: "val 4"},
				{Type: lexer.TokenAttributeListKey, Text: "key5"},
				{Type: lexer.TokenAttributeListEnd, Text: "}"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Multiple (empty) Section 1s",
			source: `
# Section 1

# Section 2
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 1"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 2"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Multiple Section 1s",
			source: `
# Section 1

Some text

# Section 2
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 1"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Some text"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 2"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Multiple Section 2s",
			source: `
# Section 1

## Section 1.1

Some text

## Section 1.2
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 1"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenSection2Begin, Text: "##"},
				{Type: lexer.TokenText, Text: "Section 1.1"},
				{Type: lexer.TokenSection2Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Some text"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenSection2End, Text: ""},
				{Type: lexer.TokenSection2Begin, Text: "##"},
				{Type: lexer.TokenText, Text: "Section 1.2"},
				{Type: lexer.TokenSection2Content, Text: ""},
				{Type: lexer.TokenSection2End, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
	}
	RunTests(t, testCases)
}

func TestLexCodeBlock(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Code Block without Language or Attributes",
			source: `
# Showcasing Code Blocks

` + "```" + `
console.log('1337')
alert('haxxed!')
` + "```" + `
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Showcasing Code Blocks"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenCodeBlockBegin, Text: "```"},
				{Type: lexer.TokenText, Text: "console.log('1337')\nalert('haxxed!')\n"},
				{Type: lexer.TokenCodeBlockEnd, Text: "```"},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Code Block with Language",
			source: `
# Showcasing Code Blocks

` + "```" + `js
console.log('1337')
alert('haxxed!')
` + "```" + `
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Showcasing Code Blocks"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenCodeBlockBegin, Text: "```"},
				{Type: lexer.TokenCodeBlockLang, Text: "js"},
				{Type: lexer.TokenText, Text: "console.log('1337')\nalert('haxxed!')\n"},
				{Type: lexer.TokenCodeBlockEnd, Text: "```"},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Code Block with Attributes",
			source: `
# Showcasing Code Blocks

` + "```" + ` {Source=https://gist.github.com/no/where Lines=1-5}
console.log('1337')
alert('haxxed!')
` + "```" + `
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Showcasing Code Blocks"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenCodeBlockBegin, Text: "```"},
				{Type: lexer.TokenAttributeListBegin, Text: "{"},
				{Type: lexer.TokenAttributeListKey, Text: "Source"},
				{Type: lexer.TokenText, Text: "https://gist.github.com/no/where"},
				{Type: lexer.TokenAttributeListKey, Text: "Lines"},
				{Type: lexer.TokenText, Text: "1-5"},
				{Type: lexer.TokenAttributeListEnd, Text: "}"},
				{Type: lexer.TokenText, Text: "console.log('1337')\nalert('haxxed!')\n"},
				{Type: lexer.TokenCodeBlockEnd, Text: "```"},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Code Block with Language and Attributes",
			source: `
# Showcasing Code Blocks

` + "```" + `js {Source=https://gist.github.com/no/where Lines=1-5}
console.log('1337')
alert('haxxed!')
` + "```" + `
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Showcasing Code Blocks"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenCodeBlockBegin, Text: "```"},
				{Type: lexer.TokenCodeBlockLang, Text: "js"},
				{Type: lexer.TokenAttributeListBegin, Text: "{"},
				{Type: lexer.TokenAttributeListKey, Text: "Source"},
				{Type: lexer.TokenText, Text: "https://gist.github.com/no/where"},
				{Type: lexer.TokenAttributeListKey, Text: "Lines"},
				{Type: lexer.TokenText, Text: "1-5"},
				{Type: lexer.TokenAttributeListEnd, Text: "}"},
				{Type: lexer.TokenText, Text: "console.log('1337')\nalert('haxxed!')\n"},
				{Type: lexer.TokenCodeBlockEnd, Text: "```"},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
	}
	RunTests(t, testCases)
}

func TestLexParagraph(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Text-only paragraph",
			source: `
# Section 1

Hello, World!
How are you doing?
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 1"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Hello, World!\nHow are you doing?\n"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Multiple text-only paragraphs",
			source: `
# Section 1

Hello, World!
How are you doing?

Good evening, Moon.
Where are you going?
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 1"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Hello, World!\nHow are you doing?"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Good evening, Moon.\nWhere are you going?\n"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Multiple paragraphs with amp specials",
			source: `
# Section 1

Hello... World!
How are you doing?

Good--evening, Moon.
Where are you going?
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 1"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Hello"},
				{Type: lexer.TokenAmpSpecial, Text: "..."},
				{Type: lexer.TokenText, Text: " World!\nHow are you doing?"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Good"},
				{Type: lexer.TokenAmpSpecial, Text: "--"},
				{Type: lexer.TokenText, Text: "evening, Moon.\nWhere are you going?\n"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Paragraph with a sidenote and sidenote definition",
			source: `
# Section 1

Hello, [世界][^1]
How are you doing?

[^1]: 世界 (Sekai) is the Japanese word for 'World'
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 1"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Hello, "},
				{Type: lexer.TokenLinkableBegin, Text: "["},
				{Type: lexer.TokenText, Text: "世界"},
				{Type: lexer.TokenSidenoteRef, Text: "1"},
				{Type: lexer.TokenLinkableEnd, Text: ""},
				{Type: lexer.TokenText, Text: "\nHow are you doing?"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenSidenoteDef, Text: "1"},
				{Type: lexer.TokenText, Text: "世界 (Sekai) is the Japanese word for 'World'"},
				//{Type: lexer.TokenText, Text: "世界 (Sekai) is the Japanese word for "},
				//{Type: lexer.TokenEnquoteSingleBegin, Text: "`"},
				//{Type: lexer.TokenText, Text: "World"},
				//{Type: lexer.TokenEnquoteSingleEnd, Text: "'"},
				{Type: lexer.TokenSidenoteDefEnd, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Paragraph with a link and link definition",
			source: `
# Section 1

Hello, [世界][1]
How are you doing?

[1]: https://jisho.org/word/世界
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 1"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Hello, "},
				{Type: lexer.TokenLinkableBegin, Text: "["},
				{Type: lexer.TokenText, Text: "世界"},
				{Type: lexer.TokenLinkRef, Text: "1"},
				{Type: lexer.TokenLinkableEnd, Text: ""},
				{Type: lexer.TokenText, Text: "\nHow are you doing?"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenLinkDef, Text: "1"},
				{Type: lexer.TokenText, Text: "https://jisho.org/word/世界"},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Paragraph with variously enquoted text",
			source: `
# Section 1

Blah blab <<blah>> blah.
Bla *blub* **bluuuhh**!!!

## Section 2

What "the" ***frick?!!***

`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 1"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Blah blab "},
				{Type: lexer.TokenEnquoteAngledBegin, Text: "<<"},
				{Type: lexer.TokenText, Text: "blah"},
				{Type: lexer.TokenEnquoteAngledEnd, Text: ">>"},
				{Type: lexer.TokenText, Text: " blah.\nBla "},
				{Type: lexer.TokenEmphasisBegin, Text: "*"},
				{Type: lexer.TokenText, Text: "blub"},
				{Type: lexer.TokenEmphasisEnd, Text: "*"},
				{Type: lexer.TokenText, Text: " "},
				{Type: lexer.TokenStrongBegin, Text: "**"},
				{Type: lexer.TokenText, Text: "bluuuhh"},
				{Type: lexer.TokenStrongEnd, Text: "**"},
				{Type: lexer.TokenText, Text: "!!!"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenSection2Begin, Text: "##"},
				{Type: lexer.TokenText, Text: "Section 2"},
				{Type: lexer.TokenSection2Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "What "},
				{Type: lexer.TokenEnquoteDoubleBegin, Text: "\""},
				{Type: lexer.TokenText, Text: "the"},
				{Type: lexer.TokenEnquoteDoubleEnd, Text: "\""},
				{Type: lexer.TokenText, Text: " "},
				{Type: lexer.TokenEmphasisStrongBegin, Text: "***"},
				{Type: lexer.TokenText, Text: "frick?!!"},
				{Type: lexer.TokenEmphasisStrongEnd, Text: "***"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenSection2End, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Paragraph with strikethrough and marked text",
			source: `
# Section 1

I ~~love~~ hate ==JavaScript==.

## Section 2

==JavaScript== is the ~~best~~ worst language.
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 1"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "I "},
				{Type: lexer.TokenStrikethroughBegin, Text: "~~"},
				{Type: lexer.TokenText, Text: "love"},
				{Type: lexer.TokenStrikethroughEnd, Text: "~~"},
				{Type: lexer.TokenText, Text: " hate "},
				{Type: lexer.TokenMarkerBegin, Text: "=="},
				{Type: lexer.TokenText, Text: "JavaScript"},
				{Type: lexer.TokenMarkerEnd, Text: "=="},
				{Type: lexer.TokenText, Text: "."},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenSection2Begin, Text: "##"},
				{Type: lexer.TokenText, Text: "Section 2"},
				{Type: lexer.TokenSection2Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenMarkerBegin, Text: "=="},
				{Type: lexer.TokenText, Text: "JavaScript"},
				{Type: lexer.TokenMarkerEnd, Text: "=="},
				{Type: lexer.TokenText, Text: " is the "},
				{Type: lexer.TokenStrikethroughBegin, Text: "~~"},
				{Type: lexer.TokenText, Text: "best"},
				{Type: lexer.TokenStrikethroughEnd, Text: "~~"},
				{Type: lexer.TokenText, Text: " worst language.\n"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenSection2End, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
	}
	RunTests(t, testCases)
}

func TestLexBlockQuote(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Blockquote without attribution",
			source: `
# Quotes and Citations

I forgot who this quote is from:
> If we stop dreaming big dreams, if we stop looking for a greater purpose,
> then we may as well be machines ourselves.
...maybe it had something to do with AI?
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Quotes and Citations"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "I forgot who this quote is from:\n"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenBlockquoteBegin, Text: ""},
				{Type: lexer.TokenText, Text: "If we stop dreaming big dreams, if we stop looking for a greater purpose,"},
				{Type: lexer.TokenText, Text: "then we may as well be machines ourselves."},
				{Type: lexer.TokenBlockquoteEnd, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenAmpSpecial, Text: "..."},
				{Type: lexer.TokenText, Text: "maybe it had something to do with AI?\n"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Blockquote with partial attribution",
			source: `
# Quotes and Citations

Oh, I remember now:

> If we stop dreaming big dreams, if we stop looking for a greater purpose,
> then we may as well be machines ourselves.
> -- Garry Kasparov
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Quotes and Citations"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Oh, I remember now:"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenBlockquoteBegin, Text: ""},
				{Type: lexer.TokenText, Text: "If we stop dreaming big dreams, if we stop looking for a greater purpose,"},
				{Type: lexer.TokenText, Text: "then we may as well be machines ourselves."},
				{Type: lexer.TokenBlockquoteAttrAuthor, Text: ""},
				{Type: lexer.TokenText, Text: "Garry Kasparov"},
				{Type: lexer.TokenBlockquoteAttrEnd, Text: ""},
				{Type: lexer.TokenBlockquoteEnd, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Blockquote with full attribution",
			source: `
# Quotes and Citations

Btw, it's from this book:

> If we stop dreaming big dreams, if we stop looking for a greater purpose,
> then we may as well be machines ourselves.
> -- Garry Kasparov, Deep Thinking
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Quotes and Citations"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Btw, it's from this book:"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenBlockquoteBegin, Text: ""},
				{Type: lexer.TokenText, Text: "If we stop dreaming big dreams, if we stop looking for a greater purpose,"},
				{Type: lexer.TokenText, Text: "then we may as well be machines ourselves."},
				{Type: lexer.TokenBlockquoteAttrAuthor, Text: ""},
				{Type: lexer.TokenText, Text: "Garry Kasparov"},
				{Type: lexer.TokenBlockquoteAttrSource, Text: ""},
				{Type: lexer.TokenText, Text: "Deep Thinking"},
				{Type: lexer.TokenBlockquoteAttrEnd, Text: ""},
				{Type: lexer.TokenBlockquoteEnd, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Blockquote containing enquote'd text",
			source: `
# Quotes and Citations

> Putting "fun" back in "fundamentally flawed."
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Quotes and Citations"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenBlockquoteBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Putting "},
				{Type: lexer.TokenEnquoteDoubleBegin, Text: "\""},
				{Type: lexer.TokenText, Text: "fun"},
				{Type: lexer.TokenEnquoteDoubleEnd, Text: "\""},
				{Type: lexer.TokenText, Text: " back in "},
				{Type: lexer.TokenEnquoteDoubleBegin, Text: "\""},
				{Type: lexer.TokenText, Text: "fundamentally flawed."},
				{Type: lexer.TokenEnquoteDoubleEnd, Text: "\""},
				{Type: lexer.TokenBlockquoteEnd, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
	}
	RunTests(t, testCases)
}

func TestLexImage(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Image without title text",
			source: `
# Image Test

Hello, here is an image:
![Image alt text](/path/to/image.png)
I hope you like it.
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Image Test"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Hello, here is an image:\n"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenImageBegin, Text: "!["},
				{Type: lexer.TokenImageAltText, Text: "Image alt text"},
				{Type: lexer.TokenImagePath, Text: "/path/to/image.png"},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "I hope you like it.\n"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Image with title text",
			source: `
# Image Test

Hello, here is an image:
![Image alt text](/path/to/image.png "Some image title")
I hope you like it.
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Image Test"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Hello, here is an image:\n"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenImageBegin, Text: "!["},
				{Type: lexer.TokenImageAltText, Text: "Image alt text"},
				{Type: lexer.TokenImagePath, Text: "/path/to/image.png"},
				{Type: lexer.TokenImageTitle, Text: "Some image title"},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "I hope you like it.\n"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
	}
	RunTests(t, testCases)
}

func TestLexEscape(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Escaped amp special",
			source: `
# Section 1

## Section 2 {#escape-tut}

The following \&nbsp; turns into a non-break space.

Here you have it in mono space: `+"`"+`&nbsp;`+"`"+`
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 1"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenSection2Begin, Text: "##"},
				{Type: lexer.TokenText, Text: "Section 2 "},
				{Type: lexer.TokenAttributeListBegin, Text: "{"},
				{Type: lexer.TokenAttributeListID, Text: "escape-tut"},
				{Type: lexer.TokenAttributeListEnd, Text: "}"},
				{Type: lexer.TokenSection2Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "The following "},
				{Type: lexer.TokenText, Text: "&"},
				{Type: lexer.TokenText, Text: "nbsp; turns into a non"},
				{Type: lexer.TokenAmpSpecial, Text: "-"},
				{Type: lexer.TokenText, Text: "break space."},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Here you have it in mono space: "},
				{Type: lexer.TokenMono, Text: "&nbsp;"},
				{Type: lexer.TokenText, Text: "\n"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenSection2End, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
	}
	RunTests(t, testCases)
}

func TestLexHtmlElement(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Html Tag at Top Level",
			source: `
<Abstract>
	Html element content
</Abstract>
`,
			expected: []lexer.Token{
				{Type: lexer.TokenHtmlTagOpen, Text: "Abstract"},
				{Type: lexer.TokenHtmlTagContent, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Html element content\n"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenHtmlTagClose, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
	}
	// @todo: needs way more tests, but the current implementation is also very broken
	RunTests(t, testCases)
}

func TestLexHorizontalRule(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Horizontal Rule (---) between two paragraphs (in Section 1)",
			source: `
# Section 1

Paragraph 1.
More text.

---

Paragraph 2.
Even more text.
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 1"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Paragraph 1.\nMore text."},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenHorizontalRule, Text: "---"},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Paragraph 2.\nEven more text.\n"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
		{
			name: "Horizontal Rule (***) between two paragraphs (in Section 2)",
			source: `
# Section 1

## Section 2

Paragraph 1.
More text.

***

Paragraph 2.
Even more text.
`,
			expected: []lexer.Token{
				{Type: lexer.TokenSection1Begin, Text: "#"},
				{Type: lexer.TokenText, Text: "Section 1"},
				{Type: lexer.TokenSection1Content, Text: ""},
				{Type: lexer.TokenSection2Begin, Text: "##"},
				{Type: lexer.TokenText, Text: "Section 2"},
				{Type: lexer.TokenSection2Content, Text: ""},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Paragraph 1.\nMore text."},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenHorizontalRule, Text: "***"},
				{Type: lexer.TokenParagraphBegin, Text: ""},
				{Type: lexer.TokenText, Text: "Paragraph 2.\nEven more text.\n"},
				{Type: lexer.TokenParagraphEnd, Text: ""},
				{Type: lexer.TokenSection2End, Text: ""},
				{Type: lexer.TokenSection1End, Text: ""},
				{Type: lexer.TokenEOF, Text: ""},
			},
		},
	}
	RunTests(t, testCases)
}

type TestCase struct {
	name, source   string
	expected       []lexer.Token
	expectedErrors []string
}

func RunTests(t *testing.T, testCases []TestCase) {
	lx := lexer.New()
	for _, testCase := range testCases {
		t.Log("now testing", testCase.name)
		lx.LexSource(testCase.name, testCase.source)
		diffTokens := deep.Equal(lx.Lexemes, testCase.expected)
		for _, diff := range diffTokens {
			t.Error(diff)
		}
		diffErrors := deep.Equal(innerErrorStrings(lx.Errors), testCase.expectedErrors)
		for _, diff := range diffErrors {
			t.Error(diff)
		}
		lx.Clear()
	}
}

func innerErrorStrings(es []error) (ss []string) {
	for _, e := range es {
		ss = append(ss, innerErrorString(e))
	}
	return ss
}

func innerErrorString(e error) string {
	if lxe, ok := e.(lexer.LexerError); ok {
		return lxe.Inner.Error()
	}
	return e.Error()
}
