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
url-path: test
rel-me: https://tech.lgbt/@attaboy
fedi-creator: @attaboy@tech.lgbt
lang: en
published: 2019-11-15
revised: 2020-05-06
est-reading: 5
series: You're oh so meta
series-prev: Lorem Ipsum
series-prev-link: /lorem
series-next: Lorem Epsum
series-next-link: /epsum
enable-revision-warning: true
---

<Abstract>
This is the abstract content.
</Abstract>

# Section 1

First section, first paragraph.
This line is still part of the same paragraph.

This line is a new paragraph,
and this is still part of it.

# Section 2

Second section, second paragraph.

## Section 2.1

Here is some code:

`+"```"+`go https://gist.github.com/cvanloo/a2801dc42ab25ddd7a0b50fe1e13ca0a L:1-7
func pointOfNoReturn(n int) (r int) {
	defer func() {
		e := recover()
		r = e + 1
	}()
	panic(n - 1)
}
`+"```"+`

...And here is the same "thing," but different:

<Code Lang="Go" Source="https://gist.github.com/cvanloo/a2801dc42ab25ddd7a0b50fe1e13ca0a#file-no_return-go-L1-L7">
func pointOfNoReturn(n int) (r int) {
	defer func() {
		e := recover()
		r = e + 1
	}()
	panic(n - 1)
}
</Code>

## Section 2.2

There is a link [here](https://example.com/), what should *I* do with it? **Click** ***it***, or what?

![Cat in a Bag](cat_in_a_bag "Image of a cat looking out of a pink bag.")

---

> かわいい
> -- Author Name, Where From
`

type expectToken struct {
	lexer.Token
	Text string
}

var testExpected = []expectToken{
	{
		Token: lexer.Token{
			Type: lexer.TokenMetaBegin,
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
			Type: lexer.TokenHtmlTagOpen,
		},
		Text: "Abstract",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraphBegin,
		},
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenText,
		},
		Text: "This is the abstract content.\n",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraphEnd,
		},
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenHtmlTagClose,
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
			Type: lexer.TokenParagraphBegin,
		},
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenText,
		},
		Text: `First section, first paragraph.
This line is still part of the same paragraph.`,
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraphEnd,
		},
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraphBegin,
		},
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenText,
		},
		Text: `This line is a new paragraph,
and so is this one.`,
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraphEnd,
		},
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenSection1,
		},
		Text: "Section 2",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraphBegin,
		},
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenText,
		},
		Text: `Second section, second paragraph.`,
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraphEnd,
		},
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenSection2,
		},
		Text: "Section 2.1",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraphBegin,
		},
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenText,
		},
		Text: `Here is some code:`,
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraphEnd,
		},
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenCodeBlockBegin,
		},
		Text: "go",
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenText,
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
			Type: lexer.TokenParagraphBegin,
		},
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenText,
		},
		Text: `...And here is the same thing, but different:`,
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenParagraphEnd,
		},
	},
	{
		Token: lexer.Token {
			Type: lexer.TokenHtmlTagOpen,
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
			Type: lexer.TokenText,
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
			Type: lexer.TokenHtmlTagClose,
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
				t.Errorf("unexpected token: %s", lx.Tokens[len(lx.Tokens)-1+i].Type)
			}
		} else {
			c := len(testExpected) - len(lx.Tokens)
			t.Errorf("lexer produced too few tokens: -%d", c)
			for i := range c {
				t.Errorf("missing token: %s", testExpected[len(testExpected)-1+i].Type)
			}
		}
	}
	for i, tok := range lx.Tokens {
		e := testExpected[i]
		if e.Type != tok.Type {
			t.Errorf("wrong token type at index: %d, expected: %s, got: %s", i, e.Type, tok.Type)
		}
		if e.Text != "" && tok.Text() != e.Text {
			t.Errorf("error at index: %d, token: %s, expected: `%s`, got: `%s`", i, tok.Type, e.Text, tok.Text())
		}
	}
}
