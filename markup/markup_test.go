package markup_test

import (
	"testing"
	"github.com/go-test/deep"
	"github.com/cvanloo/blog-go/markup/lexer"
	"github.com/cvanloo/blog-go/markup/parser"
	. "github.com/cvanloo/blog-go/assert"
	//"github.com/cvanloo/blog-go/page"
	//"github.com/cvanloo/blog-go/markup"
)

type TestCase struct {
	Comment string
	Source string
	ExpectedLexemes []lexer.Token
	ExpectedParseResult *parser.Blog
	ExpectedLexerErrors []string // @todo: nuuhuhuhuh
	ExpectedParserErrors []string
}

var TestCases = []TestCase{
	{ // @fixme: this is broken in the html, separate link definition, what if sidenote inline definition?
		Comment: "Allow for links inside sidenote content (content as separate definition).",
		Source: `
# Section 1

Some [text][^sn1] with a sidenote.

[^sn1]: This is some [text](https://example.com/hello) with a link.
`,
		ExpectedLexemes: []lexer.Token{
			{Type: lexer.TokenSection1Begin, Text: "#"},
			{Type: lexer.TokenText, Text: "Section 1"},
			{Type: lexer.TokenSection1Content, Text: ""},
			{Type: lexer.TokenParagraphBegin, Text: ""},
			{Type: lexer.TokenText, Text: "Some "},
			{Type: lexer.TokenLinkableBegin, Text: "["},
			{Type: lexer.TokenText, Text: "text"},
			{Type: lexer.TokenSidenoteRef, Text: "sn1"},
			{Type: lexer.TokenLinkableEnd, Text: ""},
			{Type: lexer.TokenText, Text: " with a sidenote."},
			{Type: lexer.TokenParagraphEnd, Text: ""},
			{Type: lexer.TokenSidenoteDef, Text: "sn1"},
			{Type: lexer.TokenText, Text: "This is some "},
			{Type: lexer.TokenLinkableBegin, Text: "["},
			{Type: lexer.TokenText, Text: "text"},
			{Type: lexer.TokenLinkHref, Text: "https://example.com/hello"},
			{Type: lexer.TokenLinkableEnd, Text: ""},
			{Type: lexer.TokenText, Text: " with a link."},
			{Type: lexer.TokenSidenoteDefEnd, Text: ""},
			{Type: lexer.TokenSection1End, Text: ""},
			{Type: lexer.TokenEOF, Text: ""},
		},
		ExpectedParseResult: &parser.Blog{
			Meta: parser.Meta{},
			Sections: []*parser.Section{
				{
					Level: 1,
					Heading: parser.TextRich{AsRef(parser.Text("Section 1"))},
					Content: []parser.Node{
						&parser.Paragraph{
							Content: []parser.Node{
								AsRef(parser.Text("Some ")),
								&parser.Sidenote{
									Ref: "sn1",
									Word: parser.TextRich{AsRef(parser.Text("text"))},
								},
								AsRef(parser.Text(" with a sidenote.")),
							},
						},
					},
				},
			},
			LinkDefinitions: map[string]string{},
			SidenoteDefinitions: map[string]parser.TextRich{
				"sn1": parser.TextRich{
					AsRef(parser.Text("This is some ")),
					&parser.Link{
						Name: parser.TextRich{AsRef(parser.Text("text"))},
						Href: "https://example.com/hello",
					},
					AsRef(parser.Text(" with a link.")),
				},
			},
			TermDefinitions: map[string]parser.TextRich{},
		},
	},
}

func TestMarkup(t *testing.T) {
	lx := lexer.New()
	for testNum, testCase := range TestCases {
		t.Log("now testing", testNum, testCase.Comment)
		lx.LexSource(testCase.Comment, testCase.Source)
		diffTokens := deep.Equal(lx.Lexemes, testCase.ExpectedLexemes)
		for _, diff := range diffTokens {
			t.Error(diff)
		}
		diffLexErrors := deep.Equal(innerErrorStrings(lx.Errors), testCase.ExpectedLexerErrors)
		for _, diff := range diffLexErrors {
			t.Error(diff)
		}
		blog, err := parser.Parse(lx)
		if err != nil {
			t.Error(err)
		}
		diffParseResult := deep.Equal(blog, testCase.ExpectedParseResult)
		for _, diff := range diffParseResult {
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
