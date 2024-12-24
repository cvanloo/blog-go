package parser_test

import (
	"testing"

	"github.com/go-test/deep"
	//"github.com/kr/pretty"

	"github.com/cvanloo/blog-go/markup"
	"github.com/cvanloo/blog-go/markup/parser"
)

func TestParsingBlog(t *testing.T) {
	blog, err := parser.Parse(markup.LexerTestTokens)
	if err != nil {
		t.Error(err)
	}
	if diff := deep.Equal(blog, markup.BlogParserTestStruct); diff != nil {
		t.Error(diff)
	}
	//t.Logf("%# v", pretty.Formatter(blog))
}

func TestParsingFixReferences(t *testing.T) {
	blog, err := parser.Parse(markup.LexerTestTokens)
	if err != nil {
		t.Error(err)
	}
	if diff := deep.Equal(blog, markup.BlogParserTestStruct); diff != nil {
		t.Fatal(diff)
	}
	refFixer := &parser.FixReferencesVisitor{}
	blog.Accept(refFixer)
	if refFixer.Errors != nil {
		t.Error(refFixer.Errors)
	}
	if diff := deep.Equal(blog, markup.BlogParserFixedTestStruct); diff != nil {
		t.Error(diff)
	}
}
