package markup

import (
	//"time"

	"github.com/cvanloo/blog-go/markup/lexer"
	"github.com/cvanloo/blog-go/markup/gen"
	//. "github.com/cvanloo/blog-go/assert"
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

const BlogTestSource = `
---
url-path: hello
title: Hello, World!
author: Colin van~Loo
---

# こんにちは、世界！

Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.

Ut enim ad minim veniam, quis nostrud---exercitation ullamco---laboris nisi ut aliquip ex ea commodo consequat.
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.

## Lorem Ipsum

Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.

## Lorem Epsum

Lorem epsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

# さようなら

Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

`

var LexerTestTokens = MockTokens{
	{Type: lexer.TokenMetaBegin, Text: "---"},
	{Type: lexer.TokenMetaKey, Text: "url-path"},
	{Type: lexer.TokenText, Text: "hello"},
	{Type: lexer.TokenMetaKey, Text: "title"},
	{Type: lexer.TokenText, Text: "Hello, World!"},
	{Type: lexer.TokenMetaKey, Text: "author"},
	{Type: lexer.TokenText, Text: "Colin van"},
	{Type: lexer.TokenAmpSpecial, Text: "~"},
	{Type: lexer.TokenText, Text: "Loo"},
	{Type: lexer.TokenMetaEnd, Text: "---"},
	{Type: lexer.TokenSection1Begin, Text: "#"},
	{Type: lexer.TokenText, Text: "こんにちは、世界！"},
	{Type: lexer.TokenSection1Content, Text: ""},
	{Type: lexer.TokenParagraphBegin, Text: ""},
	{Type: lexer.TokenText, Text: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.\nDuis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur."},
	{Type: lexer.TokenParagraphEnd, Text: ""},
	{Type: lexer.TokenParagraphBegin, Text: ""},
	{Type: lexer.TokenText, Text: "Ut enim ad minim veniam, quis nostrud"},
	{Type: lexer.TokenAmpSpecial, Text: "---"},
	{Type: lexer.TokenText, Text: "exercitation ullamco"},
	{Type: lexer.TokenAmpSpecial, Text: "---"},
	{Type: lexer.TokenText, Text: "laboris nisi ut aliquip ex ea commodo consequat.\nDuis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur."},
	{Type: lexer.TokenParagraphEnd, Text: ""},
	{Type: lexer.TokenSection2Begin, Text: "##"},
	{Type: lexer.TokenText, Text: "Lorem Ipsum"},
	{Type: lexer.TokenSection2Content, Text: ""},
	{Type: lexer.TokenParagraphBegin, Text: ""},
	{Type: lexer.TokenText, Text: "Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."},
	{Type: lexer.TokenParagraphEnd, Text: ""},
	{Type: lexer.TokenParagraphBegin, Text: ""},
	{Type: lexer.TokenText, Text: "Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur."},
	{Type: lexer.TokenParagraphEnd, Text: ""},
	{Type: lexer.TokenSection2End, Text: ""},
	{Type: lexer.TokenSection2Begin, Text: "##"},
	{Type: lexer.TokenText, Text: "Lorem Epsum"},
	{Type: lexer.TokenSection2Content, Text: ""},
	{Type: lexer.TokenParagraphBegin, Text: ""},
	{Type: lexer.TokenText, Text: "Lorem epsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."},
	{Type: lexer.TokenParagraphEnd, Text: ""},
	{Type: lexer.TokenSection2End, Text: ""},
	{Type: lexer.TokenSection1End, Text: ""},
	{Type: lexer.TokenSection1Begin, Text: "#"},
	{Type: lexer.TokenText, Text: "さようなら"},
	{Type: lexer.TokenSection1Content, Text: ""},
	{Type: lexer.TokenParagraphBegin, Text: ""},
	{Type: lexer.TokenText, Text: "Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."},
	{Type: lexer.TokenParagraphEnd, Text: ""},
	{Type: lexer.TokenSection1End, Text: ""},
	{Type: lexer.TokenEOF, Text: ""},
}

var BlogTestStruct = gen.Blog{
	UrlPath: "hello",
	Title: gen.StringOnlyContent{gen.Text("Hello, World!")},
	Author: gen.Author{
		Name: gen.StringOnlyContent{gen.Text("Colin van"), gen.AmpNoBreakSpace, gen.Text("Loo")},
	},
	Sections: []gen.Section{
		{
			Level: 1,
			Heading: gen.StringOnlyContent{gen.Text("こんにちは、世界！")},
			Content: []gen.Renderable{
				gen.Paragraph{
					Content: gen.StringOnlyContent{gen.Text("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.\nDuis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.")},
				},
				gen.Paragraph{
					Content: gen.StringOnlyContent{
						gen.Text("Ut enim ad minim veniam, quis nostrud"),
						gen.AmpEMDash,
						gen.Text("exercitation ullamco"),
						gen.AmpEMDash,
						gen.Text("laboris nisi ut aliquip ex ea commodo consequat.\nDuis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur."),
					},
				},
				gen.Section{
					Level: 2,
					Heading: gen.StringOnlyContent{gen.Text("Lorem Ipsum")},
					Content: []gen.Renderable{
						gen.Paragraph{
							Content: gen.StringOnlyContent{gen.Text("Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.")},
						},
						gen.Paragraph{
							Content: gen.StringOnlyContent{gen.Text("Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.")},
						},
					},
				},
				gen.Section{
					Level: 2,
					Heading: gen.StringOnlyContent{gen.Text("Lorem Epsum")},
					Content: []gen.Renderable{
						gen.Paragraph{
							Content: gen.StringOnlyContent{gen.Text("Lorem epsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.")},
						},
					},
				},
			},
		},
		{
			Level: 1,
			Heading: gen.StringOnlyContent{gen.Text("さようなら")},
			Content: []gen.Renderable{
				gen.Paragraph{
					Content: gen.StringOnlyContent{gen.Text("Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.")},
				},
			},
		},
	},
}
