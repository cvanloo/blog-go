package parser_test

import (
	"testing"

	"github.com/cvanloo/blog-go/markup/lexer"
	"github.com/cvanloo/blog-go/markup/parser"
	"github.com/cvanloo/blog-go/markup/gen"
)

type MockToken struct {
	lexer.Token
	MockText string
}

func (m MockToken) Text() string {
	return m.MockText
}

func M(t lexer.TokenType, s string) MockToken {
	return MockToken{
		Token: lexer.Token{
			Type: t,
		},
		MockText: s,
	}
}

func TestParsingBlog(t *testing.T) {
	tokens := []MockToken{
		M(lexer.TokenMetaBegin, "---"),
		M(lexer.TokenMetaKey, "author"),
		M(lexer.TokenText, "Colin van"),
		M(lexer.TokenAmpSpecial, "~"),
		M(lexer.TokenText, "Loo"),
		M(lexer.TokenMetaKey, "tags"),
		M(lexer.TokenMetaText, "meta test parser lexer golang"),
		M(lexer.TokenMetaKey, "template"),
		M(lexer.TokenMetaText, "blog-post"),
		M(lexer.TokenMetaKey, "title"),
		M(lexer.TokenMetaText, "This is a Test Title"),
		M(lexer.TokenMetaKey, "alt-title"),
		M(lexer.TokenMetaText, "This is a test alt-title"),
		M(lexer.TokenMetaKey, "url-path"),
		M(lexer.TokenMetaText, "test"),
		M(lexer.TokenMetaKey, "rel-me"),
		M(lexer.TokenMetaText, "https://tech.lgbt/@attaboy"),
		M(lexer.TokenMetaKey, "fedi-creator"),
		M(lexer.TokenMetaText, "@attaboy@tech.lgbt"),
		M(lexer.TokenMetaKey, "lang"),
		M(lexer.TokenMetaText, "en"),
		M(lexer.TokenMetaKey, "published"),
		M(lexer.TokenMetaText, "2019-11-15"),
		M(lexer.TokenMetaKey, "revised"),
		M(lexer.TokenMetaText, "2020-05-06"),
		M(lexer.TokenMetaKey, "est-reading"),
		M(lexer.TokenMetaText, "5"),
		M(lexer.TokenMetaKey, "series"),
		M(lexer.TokenMetaText, "You're oh so meta"),
		M(lexer.TokenMetaKey, "series-prev"),
		M(lexer.TokenMetaText, "Lorem Ipsum"),
		M(lexer.TokenMetaKey, "series-prev-link"),
		M(lexer.TokenMetaText, "/lorem"),
		M(lexer.TokenMetaKey, "series-next"),
		M(lexer.TokenMetaText, "Lorem Epsum"),
		M(lexer.TokenMetaKey, "series-next-link"),
		M(lexer.TokenMetaText, "/epsum"),
		M(lexer.TokenMetaKey, "enable-revision-warning"),
		M(lexer.TokenMetaText, "true"),
		M(lexer.TokenMetaEnd, "---"),
		M(lexer.TokenHtmlTagOpen, "Abstract"),
		M(lexer.TokenText, "This is the abstract content."),
		M(lexer.TokenHtmlTagClose, "Abstract"),
		M(lexer.TokenSection1, "# "),
		M(lexer.TokenText, "Section 1"),
		M(lexer.TokenParagraphBegin, ""),
		M(lexer.TokenText, "First section, first paragraph."),
		M(lexer.TokenText, "This line is still part of the same paragraph."),
		M(lexer.TokenParagraphEnd, ""),
		M(lexer.TokenParagraphBegin, ""),
		M(lexer.TokenText, "This line is a new paragraph,"),
		M(lexer.TokenText, "and this is still part of it."),
		M(lexer.TokenParagraphEnd, ""),
		M(lexer.TokenSection1, ""),
		M(lexer.TokenText, "Section 2"),
		M(lexer.TokenParagraphBegin, ""),
		M(lexer.TokenText, "Second section, second paragraph."),
		M(lexer.TokenParagraphEnd, ""),
		M(lexer.TokenSection2, ""),
		M(lexer.TokenText, "Section 2.1"),
		M(lexer.TokenParagraphBegin, ""),
		M(lexer.TokenText, "Here is some code:"),
		M(lexer.TokenParagraphEnd, ""),
		M(lexer.TokenCodeBlockBegin, "```"),
		M(lexer.TokenCodeBlockLang, "go"),
		M(lexer.TokenCodeBlockLineFirst, "1"),
		M(lexer.TokenCodeBlockLineLast, "7"),
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
		M(lexer.TokenParagraphEnd, ""), // @todo: how do we know if the html tag ends the paragraph or not?
		M(lexer.TokenHtmlTagOpen, "Code"),
		M(lexer.TokenHtmlTagAttrKey, "Lang"),
		M(lexer.TokenHtmlTagAttrVal, "Go"),
		M(lexer.TokenHtmlTagAttrKey, "Source"),
		M(lexer.TokenHtmlTagAttrVal, "https://gist.github.com/cvanloo/a2801dc42ab25ddd7a0b50fe1e13ca0a#file-no_return-go-L1-L7"),
		M(lexer.TokenText, "func pointOfNoReturn(n int) (r int) {"),
		M(lexer.TokenText, "\tdefer func() {"),
		M(lexer.TokenText, "\t\te := recover()"),
		M(lexer.TokenText, "\t\tr = e + 1"),
		M(lexer.TokenText, "\t}()"),
		M(lexer.TokenText, "\tpanic(n - 1)"),
		M(lexer.TokenText, "}"),
		M(lexer.TokenHtmlTagClose, "Code"),
		M(lexer.TokenSection2, ""),
		M(lexer.TokenText, "Section 2.2"),
		M(lexer.TokenParagraphBegin, ""),
		M(lexer.TokenText, "There is a link "),
		M(lexer.TokenLinkText, "here"),
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
		M(lexer.TokenImage, ""),
		M(lexer.TokenImageTitle, "Cat in a Bag"),
		M(lexer.TokenImagePath, "cat_in_a_bag"),
		M(lexer.TokenImageAlt, "Image of a cat looking out of a pink bag."), // @todo: allow for specials and stuff?
		M(lexer.TokenHorizontalRule, "---"),
		M(lexer.TokenBlockquoteBegin, ""),
		M(lexer.TokenText, "かわいい"),
		M(lexer.TokenBlockquoteAttribution, "Author Name, Where From"),
		M(lexer.TokenBlockquoteEnd, ""),
	}
	blog, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
}
