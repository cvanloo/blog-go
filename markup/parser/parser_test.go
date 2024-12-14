package parser_test

import (
	"testing"

	"github.com/go-test/deep"

	"github.com/cvanloo/blog-go/markup"
	"github.com/cvanloo/blog-go/markup/parser"
)

func TestParsingBlog(t *testing.T) {
	blog, err := parser.Parse(markup.LexerTestTokens)
	if err != nil {
		t.Error(err)
	}
	if diff := deep.Equal(blog, markup.BlogTestStruct); diff != nil {
		t.Error(diff)
	}
}
