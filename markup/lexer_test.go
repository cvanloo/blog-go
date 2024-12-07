package markup

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
</Abstract

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
	Token
	Text string
}

var testExpected = []expectToken{
	{
		Type: TokenMetaStart,
	},
	{
		Type: TokenMetaKey,
		Text: "author",
	},
	{
		Type: TokenMetaVal,
		Text: "Colin van~Loo",
	},
	{
		Type: TokenMetaKey,
		Text: "tags",
	},
	{
		Type: TokenMetaVal,
		Text: "meta test parser lexer golang",
	},
	{
		Type: TokenMetaKey,
		Text: "template",
	},
	{
		Type: TokenMetaVal,
		Text: "blog-post"
	},
	{
		Type: TokenMetaKey,
		Text: "title",
	},
	{
		Type: TokenMetaVal,
		Text: "This is a Test Title",
	},
	{
		Type: TokenMetaKey,
		Text: "alt-title",
	},
	{
		Type: TokenMetaVal,
		Text: "This is a test alt-title",
	},
	{
		Type: TokenMetaEnd,
	},

	{
		Type: HtmlTagStart,
		Text: "Abstract",
	},
	{
		Type: Paragraph,
		Text: "This is the abstract content.",
	},
	{
		Type: HtmlTagEnd,
		Text: "Abstract",
	},
	{
		Type: Section1,
		Text: "Section 1",
	},
	{
		Type: Paragraph,
		Text: `First section, first paragraph.
This line is still part of the same paragraph.
`
	},
	{
		Type: Paragraph,
		Text: `This line is a new paragraph,
and so is this one.
`
	},
	{
		Type: Section1,
		Text: "Section 2",
	},
	{
		Type: Paragraph,
		Text: `Second section, second paragraph.
`
	},
	{
		Type: Section2,
		Text: "Section 2.1",
	},
	// @todo...
}
