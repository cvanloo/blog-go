package gen_test

import (
	"fmt"
	"testing"

	"github.com/go-test/deep"
	//"github.com/kr/pretty"

	. "github.com/cvanloo/blog-go/assert"
	"github.com/cvanloo/blog-go/markup"
	"github.com/cvanloo/blog-go/markup/gen"
)

func TestGenMakeTemplateData(t *testing.T) {
	blog := markup.BlogParserFixedTestStruct
	genBlog := gen.Blog{}
	makeGen := &gen.MakeGenVisitor{
		TemplateData: &genBlog,
	}
	blog.Accept(makeGen)
	if makeGen.Errors != nil {
		t.Error(makeGen.Errors)
	}
	if diff := deep.Equal(genBlog, markup.BlogGenTestStruct); diff != nil {
		t.Error(diff)
	}
	//t.Logf("%# v", pretty.Formatter(genBlog))
}

func ExampleGenerateBlog() {
	fmt.Println(Must(gen.String(&markup.BlogGenTestStruct)))
	// Output:
	// <???>
}
