package markup

import (
	"time"

	"github.com/cvanloo/blog-go/markup/lexer"
	"github.com/cvanloo/blog-go/markup/gen"
	. "github.com/cvanloo/blog-go/assert"
)

type MockTokens []lexer.Token

func (m MockTokens) Tokens() func(func(lexer.Token) bool) {
	return func(yield func(lexer.Token) bool) {
		for _, t := range m {
			if !yield(t) {
				return
			}
		}
	}
}

func M(t lexer.TokenType, s string) lexer.Token {
	return lexer.Token{
		Type: t,
		Text: s,
	}
}

const BlogTestSource = `
---
author: Colin van~Loo
email: noreply@example.com
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
> ですね
> -- Author Name, Where From
`

var LexerTestTokens = MockTokens{
	M(lexer.TokenMetaBegin, "---"),
	M(lexer.TokenMetaKey, "author"),
	M(lexer.TokenText, "Colin van"),
	M(lexer.TokenAmpSpecial, "~"),
	M(lexer.TokenText, "Loo"),
	M(lexer.TokenMetaKey, "email"),
	M(lexer.TokenText, "noreply@example.com"),
	M(lexer.TokenMetaKey, "tags"),
	M(lexer.TokenText, "meta test parser lexer golang"),
	M(lexer.TokenMetaKey, "template"), // @todo: ignored for now
	M(lexer.TokenText, "blog-post"),
	M(lexer.TokenMetaKey, "title"),
	M(lexer.TokenText, "This is a Test Title"),
	M(lexer.TokenMetaKey, "alt-title"),
	M(lexer.TokenText, "This is a test alt-title"),
	M(lexer.TokenMetaKey, "url-path"),
	M(lexer.TokenText, "test"),
	M(lexer.TokenMetaKey, "rel-me"),
	M(lexer.TokenText, "https://tech.lgbt/@attaboy"),
	M(lexer.TokenMetaKey, "fedi-creator"),
	M(lexer.TokenText, "@attaboy@tech.lgbt"),
	M(lexer.TokenMetaKey, "lang"),
	M(lexer.TokenText, "en"),
	M(lexer.TokenMetaKey, "published"),
	M(lexer.TokenText, "2019-11-15"),
	M(lexer.TokenMetaKey, "revised"),
	M(lexer.TokenText, "2020-05-06"),
	M(lexer.TokenMetaKey, "est-reading"),
	M(lexer.TokenText, "5"),
	M(lexer.TokenMetaKey, "series"), // @todo: ignored for now
	M(lexer.TokenText, "You're oh so meta"),
	M(lexer.TokenMetaKey, "series-prev"),
	M(lexer.TokenText, "Lorem Ipsum"),
	M(lexer.TokenMetaKey, "series-prev-link"),
	M(lexer.TokenText, "/lorem"),
	M(lexer.TokenMetaKey, "series-next"),
	M(lexer.TokenText, "Lorem Epsum"),
	M(lexer.TokenMetaKey, "series-next-link"),
	M(lexer.TokenText, "/epsum"),
	M(lexer.TokenMetaKey, "enable-revision-warning"),
	M(lexer.TokenText, "true"),
	M(lexer.TokenMetaEnd, "---"),
	M(lexer.TokenHtmlTagOpen, "Abstract"),
	M(lexer.TokenHtmlTagContent, ""),
	M(lexer.TokenText, "This is the abstract content."),
	M(lexer.TokenHtmlTagClose, "Abstract"),
	M(lexer.TokenSection1Begin, "# "),
	M(lexer.TokenText, "Section 1"),
	M(lexer.TokenSection1Content, ""),
	M(lexer.TokenParagraphBegin, ""),
	M(lexer.TokenText, "First section, first paragraph."),
	M(lexer.TokenText, "This line is still part of the same paragraph."),
	M(lexer.TokenParagraphEnd, ""),
	M(lexer.TokenParagraphBegin, ""),
	M(lexer.TokenText, "This line is a new paragraph,"),
	M(lexer.TokenText, "and this is still part of it."),
	M(lexer.TokenParagraphEnd, ""),
	M(lexer.TokenSection1End, ""),
	M(lexer.TokenSection1Begin, ""),
	M(lexer.TokenText, "Section 2"),
	M(lexer.TokenSection1Content, ""),
	M(lexer.TokenParagraphBegin, ""),
	M(lexer.TokenText, "Second section, second paragraph."),
	M(lexer.TokenParagraphEnd, ""),
	M(lexer.TokenSection2Begin, ""),
	M(lexer.TokenText, "Section 2.1"),
	M(lexer.TokenSection2Content, ""),
	M(lexer.TokenParagraphBegin, ""),
	M(lexer.TokenText, "Here is some code:"),
	M(lexer.TokenParagraphEnd, ""),
	M(lexer.TokenCodeBlockBegin, "```"),
	M(lexer.TokenCodeBlockLang, "go"), // @todo: ignored for now
	M(lexer.TokenCodeBlockLineFirst, "1"), // @todo: ignored for now
	M(lexer.TokenCodeBlockLineLast, "7"), // @todo: ignored for now
	M(lexer.TokenText, "func pointOfNoReturn(n int) (r int) {"),
	M(lexer.TokenText, "\tdefer func() {"),
	M(lexer.TokenText, "\t\te := recover()"),
	M(lexer.TokenText, "\t\tr = e + 1"),
	M(lexer.TokenText, "\t}()"),
	M(lexer.TokenText, "\tpanic(n - 1)"),
	M(lexer.TokenText, "}"),
	M(lexer.TokenCodeBlockEnd, "```"),
	M(lexer.TokenParagraphBegin, ""),
	M(lexer.TokenAmpSpecial, "..."),
	M(lexer.TokenText, "And here is the same "),
	M(lexer.TokenEnquoteBegin, "\""),
	M(lexer.TokenText, "thing,"),
	M(lexer.TokenEnquoteEnd, "\""),
	M(lexer.TokenText, "but different:"),
	M(lexer.TokenParagraphEnd, ""),
	M(lexer.TokenHtmlTagOpen, "Code"),
	M(lexer.TokenHtmlTagAttrKey, "Lang"),
	M(lexer.TokenHtmlTagAttrVal, "Go"),
	M(lexer.TokenHtmlTagAttrKey, "Source"),
	M(lexer.TokenHtmlTagAttrVal, "https://gist.github.com/cvanloo/a2801dc42ab25ddd7a0b50fe1e13ca0a#file-no_return-go-L1-L7"),
	M(lexer.TokenHtmlTagContent, ""),
	M(lexer.TokenText, "func pointOfNoReturn(n int) (r int) {"),
	M(lexer.TokenText, "\tdefer func() {"),
	M(lexer.TokenText, "\t\te := recover()"),
	M(lexer.TokenText, "\t\tr = e + 1"),
	M(lexer.TokenText, "\t}()"),
	M(lexer.TokenText, "\tpanic(n - 1)"),
	M(lexer.TokenText, "}"),
	M(lexer.TokenHtmlTagClose, "Code"),
	M(lexer.TokenSection2End, ""),
	M(lexer.TokenSection2Begin, ""),
	M(lexer.TokenText, "Section 2.2"),
	M(lexer.TokenSection2Content, ""),
	M(lexer.TokenParagraphBegin, ""),
	M(lexer.TokenText, "There is a link "),
	M(lexer.TokenLink, ""),
	M(lexer.TokenText, "here"),
	M(lexer.TokenLinkHref, "https://example.com/"),
	M(lexer.TokenText, ", what should "),
	M(lexer.TokenEmphasis, "*"),
	M(lexer.TokenText, "I"),
	M(lexer.TokenEmphasis, "*"),
	M(lexer.TokenText, " do with it? "),
	M(lexer.TokenStrong, "**"),
	M(lexer.TokenText, "Click"),
	M(lexer.TokenStrong, "**"),
	M(lexer.TokenText, " "),
	M(lexer.TokenEmphasisStrong, "***"),
	M(lexer.TokenText, "it"),
	M(lexer.TokenEmphasisStrong, "***"),
	M(lexer.TokenText, ", or what?"),
	M(lexer.TokenParagraphEnd, ""),
	M(lexer.TokenImageBegin, ""),
	M(lexer.TokenImageTitle, ""),
	M(lexer.TokenText, "Cat in a Bag"),
	M(lexer.TokenImageAttrEnd, ""),
	M(lexer.TokenImagePath, "cat_in_a_bag"),
	M(lexer.TokenImageAlt, ""),
	M(lexer.TokenText, "Image of a cat looking out of a pink bag."),
	M(lexer.TokenImageAttrEnd, ""),
	M(lexer.TokenImageEnd, ""),
	M(lexer.TokenHorizontalRule, "---"),
	M(lexer.TokenBlockquoteBegin, ""),
	M(lexer.TokenText, "かわいい"),
	M(lexer.TokenText, "ですね"),
	M(lexer.TokenBlockquoteAttrAuthor, ""),
	M(lexer.TokenText, "Author Name"),
	M(lexer.TokenBlockquoteAttrEnd, ""),
	M(lexer.TokenBlockquoteAttrSource, ""),
	M(lexer.TokenText, "Where From"),
	M(lexer.TokenBlockquoteAttrEnd, ""),
	M(lexer.TokenBlockquoteEnd, ""),
	M(lexer.TokenSection2End, ""),
	M(lexer.TokenSection1End, ""),
	M(lexer.TokenEOF, ""),
}

var BlogTestStruct = gen.Blog{
	UrlPath: "test",
	Author: gen.Author{
		Name: gen.StringOnlyContent{
			gen.Text("Colin van"),
			gen.NoBreakSpace,
			gen.Text("Loo"),
		},
		Email: gen.StringOnlyContent{gen.Text("noreply@example.com")},
		RelMe: gen.StringOnlyContent{gen.Text("https://tech.lgbt/@attaboy")},
		FediCreator: gen.StringOnlyContent{gen.Text("@attaboy@tech.lgbt")},
	},
	Lang: "en",
	Title: gen.StringOnlyContent{gen.Text("This is a Test Title")},
	AltTitle: gen.StringOnlyContent{gen.Text("This is a test alt-title")},
	Published: gen.Revision{
		Published: Must(time.Parse("2006-01-02", "2019-11-15")),
		Revised: AsRef(Must(time.Parse("2006-01-02", "2020-05-06"))),
	},
	EstReading: 5,
	Tags: []gen.Tag{
		gen.Tag("meta"),
		gen.Tag("test"),
		gen.Tag("parser"),
		gen.Tag("lexer"),
		gen.Tag("golang"),
	},
	Series: &gen.Series{
		Prev: &gen.SeriesItem{
			Title: gen.StringOnlyContent{gen.Text("Lorem Ipsum")},
			Link: "/lorem",
		},
		Next: &gen.SeriesItem{
			Title: gen.StringOnlyContent{gen.Text("Lorem Epsum")},
			Link: "/epsum",
		},
	},
	EnableRevisionWarning: true,
	//TOC: nil, @todo: compute based on blog.Sections?
	Abstract: gen.StringOnlyContent{gen.Text("This is the abstract content.")},
	Sections: []gen.Section{
		{
			Level: 1,
			Heading: gen.StringOnlyContent{gen.Text("Section 1")},
			Content: []gen.Renderable{
				gen.Paragraph{
					Content: gen.StringOnlyContent{
						gen.Text("First section, first paragraph."),
						gen.Text("This line is still part of the same paragraph."),
					},
				},
				gen.Paragraph{
					Content: gen.StringOnlyContent{
						gen.Text("This line is a new paragraph,"),
						gen.Text("and this is still part of it."),
					},
				},
			},
		},
		{
			Level: 1, // @todo: compute level instead of storing in struct?
			Heading: gen.StringOnlyContent{gen.Text("Section 2")},
			Content: []gen.Renderable{
				gen.Paragraph{
					Content: gen.StringOnlyContent{
						gen.Text("Second section, second paragraph."),
					},
				},
				gen.Section{
					Level: 2,
					Heading: gen.StringOnlyContent{gen.Text("Section 2.1")},
					Content: []gen.Renderable{
						gen.Paragraph{
							Content: gen.StringOnlyContent{
								gen.Text("Here is some code:"),
							},
						},
						gen.CodeBlock{
							Lines: []string{
								"func pointOfNoReturn(n int) (r int) {",
								"\tdefer func() {",
								"\t\te := recover()",
								"\t\tr = e + 1",
								"\t}()",
								"\tpanic(n - 1)",
								"}",
							},
						},
						gen.Paragraph{
							Content: gen.StringOnlyContent{
								gen.Ellipsis,
								gen.Text("And here is the same "),
								gen.Enquote{gen.StringOnlyContent{gen.Text("thing,")}},
								gen.Text("but different:"),
							},
						},
						gen.CodeBlock{
							Lines: []string{
								"func pointOfNoReturn(n int) (r int) {",
								"\tdefer func() {",
								"\t\te := recover()",
								"\t\tr = e + 1",
								"\t}()",
								"\tpanic(n - 1)",
								"}",
							},
						},
					},
				},
				gen.Section{
					Level: 2,
					Heading: gen.StringOnlyContent{gen.Text("Section 2.2")},
					Content: []gen.Renderable{
						gen.Paragraph{
							Content: gen.StringOnlyContent{
								gen.Text("There is a link "),
								gen.Link{
									Href: "https://example.com/",
									Name: gen.StringOnlyContent{gen.Text("here")},
								},
								gen.Text(", what should "),
								gen.Emphasis{
									gen.StringOnlyContent{
										gen.Text("I"),
									},
								},
								gen.Text(" do with it? "),
								gen.Strong{
									gen.StringOnlyContent{
										gen.Text("Click"),
									},
								},
								gen.Text(" "),
								gen.EmphasisStrong{
									gen.StringOnlyContent{
										gen.Text("it"),
									},
								},
								gen.Text(", or what?"),
							},
						},
						gen.Image{
							Name: "cat_in_a_bag",
							Title: gen.StringOnlyContent{gen.Text("Cat in a Bag")},
							Alt: gen.StringOnlyContent{gen.Text("Image of a cat looking out of a pink bag.")},
						},
						gen.HorizontalRule{},
						gen.Blockquote{
							QuoteText: gen.StringOnlyContent{
								gen.Text("かわいい"),
								gen.Text("ですね"),
							},
							Author: gen.StringOnlyContent{gen.Text("Author Name")}, // @todo: how to split this? shouldn't we already do this in the lexer?
							Source: gen.StringOnlyContent{gen.Text("Where From")},
						},
					},
				},
			},
		},
	},
	//Relevant *RelevantBox @todo
}
