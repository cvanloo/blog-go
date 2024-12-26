package markup

import (
	//"time"

	"github.com/cvanloo/blog-go/markup/lexer"
	"github.com/cvanloo/blog-go/markup/parser"
	"github.com/cvanloo/blog-go/page"
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

const BlogTestSource = `
---
url-path: hello
title: Hello, World!
author: Colin van~Loo
lang: en
---

# こんにちは、世界！ {#s1}

Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.

Ut enim ad minim veniam, quis nostrud---exercitation ullamco---laboris nisi ut aliquip ex ea commodo consequat.
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.

## Lorem Ipsum

Ut enim ad minim [veniam](https://example.com/), quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.

## Lorem Epsum {#s2.2}

Lorem [epsum][^1] dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

[^1]: See what I did there?

# さようなら

Ut enim ad minim [veniam][0], quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

[0]: https://example.com/

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
	{Type: lexer.TokenMetaKey, Text: "lang"},
	{Type: lexer.TokenText, Text: "en"},
	{Type: lexer.TokenMetaEnd, Text: "---"},
	{Type: lexer.TokenSection1Begin, Text: "#"},
	{Type: lexer.TokenText, Text: "こんにちは、世界！ "},
	{Type: lexer.TokenAttributeListBegin, Text: "{"},
	{Type: lexer.TokenAttributeListID, Text: "s1"},
	{Type: lexer.TokenAttributeListEnd, Text: "}"},
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
	{Type: lexer.TokenText, Text: "Ut enim ad minim "},
	{Type: lexer.TokenLinkableBegin, Text: "["},
	{Type: lexer.TokenText, Text: "veniam"},
	{Type: lexer.TokenLinkHref, Text: "https://example.com/"},
	{Type: lexer.TokenLinkableEnd, Text: ""},
	{Type: lexer.TokenText, Text: ", quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."},
	{Type: lexer.TokenParagraphEnd, Text: ""},
	{Type: lexer.TokenParagraphBegin, Text: ""},
	{Type: lexer.TokenText, Text: "Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur."},
	{Type: lexer.TokenParagraphEnd, Text: ""},
	{Type: lexer.TokenSection2End, Text: ""},
	{Type: lexer.TokenSection2Begin, Text: "##"},
	{Type: lexer.TokenText, Text: "Lorem Epsum "},
	{Type: lexer.TokenAttributeListBegin, Text: "{"},
	{Type: lexer.TokenAttributeListID, Text: "s2.2"},
	{Type: lexer.TokenAttributeListEnd, Text: "}"},
	{Type: lexer.TokenSection2Content, Text: ""},
	{Type: lexer.TokenParagraphBegin, Text: ""},
	{Type: lexer.TokenText, Text: "Lorem "},
	{Type: lexer.TokenLinkableBegin, Text: "["},
	{Type: lexer.TokenText, Text: "epsum"},
	{Type: lexer.TokenSidenoteRef, Text: "1"},
	{Type: lexer.TokenLinkableEnd, Text: ""},
	{Type: lexer.TokenText, Text: " dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."},
	{Type: lexer.TokenParagraphEnd, Text: ""},
	{Type: lexer.TokenSidenoteDef, Text: "1"},
	{Type: lexer.TokenText, Text: "See what I did there?"},
	{Type: lexer.TokenSidenoteDefEnd, Text: ""},
	{Type: lexer.TokenSection2End, Text: ""},
	{Type: lexer.TokenSection1End, Text: ""},
	{Type: lexer.TokenSection1Begin, Text: "#"},
	{Type: lexer.TokenText, Text: "さようなら"},
	{Type: lexer.TokenSection1Content, Text: ""},
	{Type: lexer.TokenParagraphBegin, Text: ""},
	{Type: lexer.TokenText, Text: "Ut enim ad minim "},
	{Type: lexer.TokenLinkableBegin, Text: "["},
	{Type: lexer.TokenText, Text: "veniam"},
	{Type: lexer.TokenLinkRef, Text: "0"},
	{Type: lexer.TokenLinkableEnd, Text: ""},
	{Type: lexer.TokenText, Text: ", quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."},
	{Type: lexer.TokenParagraphEnd, Text: ""},
	{Type: lexer.TokenLinkDef, Text: "0"},
	{Type: lexer.TokenText, Text: "https://example.com/"},
	{Type: lexer.TokenSection1End, Text: ""},
	{Type: lexer.TokenEOF, Text: ""},
}

var BlogParserTestStruct = &parser.Blog{
	Meta: parser.Meta{
		"url-path": []parser.TextSimple{
			[]parser.Node{
				AsRef(parser.Text("hello")),
			},
		},
		"title": []parser.TextSimple{
			[]parser.Node{
				AsRef(parser.Text("Hello, World!")),
			},
		},
		"author": []parser.TextSimple{
			[]parser.Node{
				AsRef(parser.Text("Colin van")),
				AsRef(parser.AmpSpecial("~")),
				AsRef(parser.Text("Loo")),
			},
		},
		"lang": []parser.TextSimple{
			[]parser.Node{
				AsRef(parser.Text("en")),
			},
		},
	},
	Sections: []*parser.Section{
		{
			Level: 1,
			Attributes: parser.Attributes{
				"id": "s1",
			},
			Heading: parser.TextRich{
				AsRef(parser.Text("こんにちは、世界！ ")),
			},
			Content: []parser.Node{
				&parser.Paragraph{
					Content: []parser.Node{
						AsRef(parser.Text("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.\nDuis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.")),
					},
				},
				&parser.Paragraph{
					Content: []parser.Node{
						AsRef(parser.Text("Ut enim ad minim veniam, quis nostrud")),
						AsRef(parser.AmpSpecial("---")),
						AsRef(parser.Text("exercitation ullamco")),
						AsRef(parser.AmpSpecial("---")),
						AsRef(parser.Text("laboris nisi ut aliquip ex ea commodo consequat.\nDuis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.")),
					},
				},
				&parser.Section{
					Level: 2,
					Heading: parser.TextRich{
						AsRef(parser.Text("Lorem Ipsum")),
					},
					Content: []parser.Node{
						&parser.Paragraph{
							Content: []parser.Node{
								AsRef(parser.Text("Ut enim ad minim ")),
								&parser.Link{
									Name: parser.TextRich{AsRef(parser.Text("veniam"))},
									Href: "https://example.com/",
								},
								AsRef(parser.Text(", quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.")),
							},
						},
						&parser.Paragraph{
							Content: []parser.Node{
								AsRef(parser.Text("Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.")),
							},
						},
					},
				},
				&parser.Section{
					Level: 2,
					Attributes: parser.Attributes{
						"id": "s2.2",
					},
					Heading: parser.TextRich{
						AsRef(parser.Text("Lorem Epsum ")),
					},
					Content: []parser.Node{
						&parser.Paragraph{
							Content: []parser.Node{
								AsRef(parser.Text("Lorem ")),
								&parser.Sidenote{
									Ref: "1",
									Word: parser.TextRich{AsRef(parser.Text("epsum"))},
								},
								AsRef(parser.Text(" dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.")),
							},
						},
					},
				},
			},
		},
		{
			Level: 1,
			Heading: parser.TextRich{
				AsRef(parser.Text("さようなら")),
			},
			Content: []parser.Node{
				&parser.Paragraph{
					Content: []parser.Node{
						AsRef(parser.Text("Ut enim ad minim ")),
						&parser.Link{
							Ref: "0",
							Name: parser.TextRich{AsRef(parser.Text("veniam"))},
						},
						AsRef(parser.Text(", quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.")),
					},
				},
			},
		},
	},
	LinkDefinitions: map[string]string{
		"0": "https://example.com/",
	},
	SidenoteDefinitions: map[string]parser.TextRich{
		"1": parser.TextRich{AsRef(parser.Text("See what I did there?"))},
	},
	TermDefinitions: map[string]parser.TextRich{},
}

var BlogParserFixedTestStruct = &parser.Blog{
	Meta: parser.Meta{
		"url-path": []parser.TextSimple{
			[]parser.Node{
				AsRef(parser.Text("hello")),
			},
		},
		"title": []parser.TextSimple{
			[]parser.Node{
				AsRef(parser.Text("Hello, World!")),
			},
		},
		"author": []parser.TextSimple{
			[]parser.Node{
				AsRef(parser.Text("Colin van")),
				AsRef(parser.AmpSpecial("~")),
				AsRef(parser.Text("Loo")),
			},
		},
		"lang": []parser.TextSimple{
			[]parser.Node{
				AsRef(parser.Text("en")),
			},
		},
	},
	Sections: []*parser.Section{
		{
			Level: 1,
			Attributes: parser.Attributes{
				"id": "s1",
			},
			Heading: parser.TextRich{
				AsRef(parser.Text("こんにちは、世界！ ")),
			},
			Content: []parser.Node{
				&parser.Paragraph{
					Content: []parser.Node{
						AsRef(parser.Text("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.\nDuis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.")),
					},
				},
				&parser.Paragraph{
					Content: []parser.Node{
						AsRef(parser.Text("Ut enim ad minim veniam, quis nostrud")),
						AsRef(parser.AmpSpecial("---")),
						AsRef(parser.Text("exercitation ullamco")),
						AsRef(parser.AmpSpecial("---")),
						AsRef(parser.Text("laboris nisi ut aliquip ex ea commodo consequat.\nDuis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.")),
					},
				},
				&parser.Section{
					Level: 2,
					Heading: parser.TextRich{
						AsRef(parser.Text("Lorem Ipsum")),
					},
					Content: []parser.Node{
						&parser.Paragraph{
							Content: []parser.Node{
								AsRef(parser.Text("Ut enim ad minim ")),
								&parser.Link{
									Name: parser.TextRich{AsRef(parser.Text("veniam"))},
									Href: "https://example.com/",
								},
								AsRef(parser.Text(", quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.")),
							},
						},
						&parser.Paragraph{
							Content: []parser.Node{
								AsRef(parser.Text("Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.")),
							},
						},
					},
				},
				&parser.Section{
					Level: 2,
					Attributes: parser.Attributes{
						"id": "s2.2",
					},
					Heading: parser.TextRich{
						AsRef(parser.Text("Lorem Epsum ")),
					},
					Content: []parser.Node{
						&parser.Paragraph{
							Content: []parser.Node{
								AsRef(parser.Text("Lorem ")),
								&parser.Sidenote{
									Ref: "1",
									Word: parser.TextRich{AsRef(parser.Text("epsum"))},
									Content: parser.TextRich{AsRef(parser.Text("See what I did there?"))},
								},
								AsRef(parser.Text(" dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.")),
							},
						},
					},
				},
			},
		},
		{
			Level: 1,
			Heading: parser.TextRich{
				AsRef(parser.Text("さようなら")),
			},
			Content: []parser.Node{
				&parser.Paragraph{
					Content: []parser.Node{
						AsRef(parser.Text("Ut enim ad minim ")),
						&parser.Link{
							Ref: "0",
							Name: parser.TextRich{AsRef(parser.Text("veniam"))},
							Href: "https://example.com/",
						},
						AsRef(parser.Text(", quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.")),
					},
				},
			},
		},
	},
	LinkDefinitions: map[string]string{
		"0": "https://example.com/",
	},
	SidenoteDefinitions: map[string]parser.TextRich{
		"1": parser.TextRich{AsRef(parser.Text("See what I did there?"))},
	},
	TermDefinitions: map[string]parser.TextRich{},
}

var BlogGenTestStruct = page.Post{
	UrlPath: "hello",
	Title: page.StringOnlyContent{page.Text("Hello, World!")},
	Author: page.Author{
		Name: page.StringOnlyContent{page.Text("Colin van"), page.AmpNoBreakSpace, page.Text("Loo")},
	},
	Lang: "en",
	TOC: page.TableOfContents{
		Sections: []page.TOCSection{
			{
				ID: "s1",
				Heading: page.StringOnlyContent{page.Text("こんにちは、世界！ ")},
				NextLevel: []page.TOCSection{
					{
						ID: "lorem-ipsum",
						Heading: page.StringOnlyContent{page.Text("Lorem Ipsum")},
					},
					{
						ID: "s2.2",
						Heading: page.StringOnlyContent{page.Text("Lorem Epsum ")},
					},
				},
			},
			{
				ID: "さようなら",
				Heading: page.StringOnlyContent{page.Text("さようなら")},
			},
		},
	},
	Sections: []page.Section{
		{
			Level: 1,
			Attributes: map[string]string{
				"id": "s1",
			},
			Heading: page.StringOnlyContent{page.Text("こんにちは、世界！ ")},
			Content: []page.Renderable{
				page.Paragraph{
					Content: page.StringOnlyContent{page.Text("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.\nDuis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.")},
				},
				page.Paragraph{
					Content: page.StringOnlyContent{
						page.Text("Ut enim ad minim veniam, quis nostrud"),
						page.AmpEmDash,
						page.Text("exercitation ullamco"),
						page.AmpEmDash,
						page.Text("laboris nisi ut aliquip ex ea commodo consequat.\nDuis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur."),
					},
				},
				page.Section{
					Level: 2,
					Heading: page.StringOnlyContent{page.Text("Lorem Ipsum")},
					Content: []page.Renderable{
						page.Paragraph{
							Content: page.StringOnlyContent{
								page.Text("Ut enim ad minim "),
								page.Link{
									Href: "https://example.com/",
									Name: page.StringOnlyContent{page.Text("veniam")},
								},
								page.Text(", quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."),
							},
						},
						page.Paragraph{
							Content: page.StringOnlyContent{page.Text("Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.")},
						},
					},
				},
				page.Section{
					Level: 2,
					Attributes: map[string]string{
						"id": "s2.2",
					},
					Heading: page.StringOnlyContent{page.Text("Lorem Epsum ")},
					Content: []page.Renderable{
						page.Paragraph{
							Content: page.StringOnlyContent{
								page.Text("Lorem "),
								page.Sidenote{
									Word: page.StringOnlyContent{page.Text("epsum")},
									Content: page.StringOnlyContent{page.Text("See what I did there?")},
								},
								page.Text(" dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\nUt enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."),
							},
						},
					},
				},
			},
		},
		{
			Level: 1,
			Heading: page.StringOnlyContent{page.Text("さようなら")},
			Content: []page.Renderable{
				page.Paragraph{
					Content: page.StringOnlyContent{
						page.Text("Ut enim ad minim "),
						page.Link{
							Name: page.StringOnlyContent{page.Text("veniam")},
							Href: "https://example.com/",
						},
						page.Text(", quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."),
					},
				},
			},
		},
	},
}
