package page_test

import (
	"testing"

	"github.com/go-test/deep"
	//"github.com/kr/pretty"

	"github.com/cvanloo/blog-go/markup"
	"github.com/cvanloo/blog-go/page"
)

func TestGenMakeTemplateData(t *testing.T) {
	blog := markup.BlogParserFixedTestStruct
	post := page.Post{}
	makeGen := &page.MakeGenVisitor{
		TemplateData: &post,
	}
	blog.Accept(makeGen)
	if makeGen.Errors != nil {
		t.Error(makeGen.Errors)
	}
	if diff := deep.Equal(post, markup.BlogGenTestStruct); diff != nil {
		t.Error(diff)
	}
	//t.Logf("%# v", pretty.Formatter(genBlog))
}
