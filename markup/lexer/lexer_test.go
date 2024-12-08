package lexer_test

import (
	"testing"

	"github.com/cvanloo/blog-go/markup/lexer"
)

const testBlog = `
---
author: Colin van~Loo
tags: meta test parser lexer golang
template: blog-post
title: This is a Test Title
alt-title: This is a test alt-title
---

<Abstract>
This is the abstract content.
</Abstract>

# Section 1

First section, first paragraph.
This line is still part of the same paragraph.

This line is a new paragraph,
and so is this one.

# Section 2

Second section, second paragraph.

## Section 2.1

Here is some code:

`+"```"+`go
func pointOfNoReturn(n int) (r int) {
	defer func() {
		e := recover()
		r = e + 1
	}()
	panic(n - 1)
}
`+"```"+`

...And here is the same thing, but different:

<Code Lang="Go" Source="gist.github.com/this-link-doesnt-actually-exist#L39-95">
func pointOfNoReturn(n int) (r int) {
	defer func() {
		e := recover()
		r = e + 1
	}()
	panic(n - 1)
}
</Code>
`

type expectToken struct {
	lexer.Token
	Text string
}

var testExpected = []expectToken{
	{
		Token: lexer.Token{
			Type: lexer.TokenMetaStart,
		},
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenMetaKey,
		},
		Text: "author",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenMetaVal,
		},
		Text: "Colin van~Loo",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenMetaKey,
		},
		Text: "tags",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenMetaVal,
		},
		Text: "meta test parser lexer golang",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenMetaKey,
		},
		Text: "template",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenMetaVal,
		},
		Text: "blog-post",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenMetaKey,
		},
		Text: "title",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenMetaVal,
		},
		Text: "This is a Test Title",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenMetaKey,
		},
		Text: "alt-title",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenMetaVal,
		},
		Text: "This is a test alt-title",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenMetaEnd,
		},
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenHtmlTagStart,
		},
		Text: "Abstract",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraph,
		},
		Text: "This is the abstract content.",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenHtmlTagEnd,
		},
		Text: "Abstract",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenSection1,
		},
		Text: "Section 1",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraph,
		},
		Text: `First section, first paragraph.
This line is still part of the same paragraph.
`,
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraph,
		},
		Text: `This line is a new paragraph,
and so is this one.
`,
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenSection1,
		},
		Text: "Section 2",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraph,
		},
		Text: `Second section, second paragraph.
`,
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenSection2,
		},
		Text: "Section 2.1",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraph,
		},
		Text: `Here is some code:
`,
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenCodeBlockStart,
		},
		Text: "go",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraph,
		},
		Text: `func pointOfNoReturn(n int) (r int) {
	defer func() {
		e := recover()
		r = e + 1
	}()
	panic(n - 1)
}
`,
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenCodeBlockEnd,
		},
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraph,
		},
		Text: `...And here is the same thing, but different:
`,
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenHtmlTagStart,
		},
		Text: "Code",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenHtmlTagAttrKey,
		},
		Text: "Lang",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenHtmlTagAttrVal,
		},
		Text: "Go",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenHtmlTagAttrKey,
		},
		Text: "Source",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenHtmlTagAttrVal,
		},
		Text: "gist.github.com/this-link-doesnt-actually-exist#L39-95",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraph,
		},
		Text: `func pointOfNoReturn(n int) (r int) {
	defer func() {
		e := recover()
		r = e + 1
	}()
	panic(n - 1)
}
`,
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenHtmlTagEnd,
		},
		Text: "Code",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenEOF,
		},
	},
}

func TestLexer(t *testing.T) {
	lx := lexer.New()
	if err := lx.LexSource("testBlog", testBlog); err != nil {
		for _, tok := range lx.Tokens {
			t.Log(tok)
		}
		t.Errorf("got %d errors, first error is: %s", len(lx.Errors), err)
		for i, err := range lx.Errors {
			t.Errorf("error %d: %s", i, err)
		}
		t.FailNow()
	}
	if len(lx.Tokens) != len(testExpected) {
		t.Errorf("expected %d tokens, got %d tokens", len(lx.Tokens), len(testExpected))
		l := min(len(lx.Tokens), len(testExpected))
		for i := range l {
			if lx.Tokens[i].Type != testExpected[i].Type {
				t.Errorf("tokens don't match at index: %d, expected: %s, got: %s", i, testExpected[i].Type, lx.Tokens[i].Type)
			}
		}
		if len(lx.Tokens) > len(testExpected) {
			c := len(lx.Tokens) - len(testExpected)
			t.Errorf("lexer produced too many tokens: +%d", c)
			for i := range c {
				t.Errorf("unexpected token: %s", lx.Tokens[len(lx.Tokens)-1+i])
			}
		} else {
			c := len(testExpected) - len(lx.Tokens)
			t.Errorf("lexer produced too few tokens: -%d", c)
			for i := range c {
				t.Errorf("missing token: %s", testExpected[len(testExpected)-1+i])
			}
		}
	}
	for i, tok := range lx.Tokens {
		e := testExpected[i]
		if e.Type != tok.Type {
			t.Errorf("wrong token type at index: %d, expected: %s, got: %s", i, e.Type, tok.Type)
		}
		if e.Text != "" && tok.Text() != e.Text {
			t.Errorf("error at index: %d, token: %s, expected: %s, got: %s", i, tok.Type, e.Text, tok.Text())
		}
	}
}
