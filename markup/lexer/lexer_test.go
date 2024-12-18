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
	var metaTests = []struct{
		name, source string
		expected []lexer.Token
		expectedErrors []error
	}{
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
			// @todo: how are we going to verify the error?
			//expectedErrors: []error{
			//  "expected ---"
			//},
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
	}
	lx := lexer.New()
	for _, testCase := range metaTests {
		t.Log("now testing", testCase.name)
		lx.LexSource(testCase.name, testCase.source)
		diffTokens := deep.Equal(lx.Lexemes, testCase.expected)
		for _, diff := range diffTokens {
			t.Error(diff)
		}
		diffErrors := deep.Equal(lx.Errors, testCase.expectedErrors)
		for _, diff := range diffErrors {
			t.Error(diff)
		}
		lx.Clear()
	}
}
