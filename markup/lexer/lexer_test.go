package lexer_test

import (
	"testing"

	"github.com/cvanloo/blog-go/markup"
	"github.com/cvanloo/blog-go/markup/lexer"
)

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
}
