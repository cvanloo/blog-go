package gen_test

import (
	"fmt"

	. "github.com/cvanloo/blog-go/assert"
	"github.com/cvanloo/blog-go/markup"
	"github.com/cvanloo/blog-go/markup/gen"
)

func TestGenMakeTemplateData(t *testing.T) {
	blog := markup.BlogParserFixedTestStruct
	makeGen := gen.MakeGenVisitor{}
	blog.Accept(makeGen)
	if makeGen.Errors != nil {
		t.Error(makeGen.Errors)
	}
	if diff := deep.Equal(blog, markup.BlogGenTestStruct); diff != nil {
		t.Error(diff)
	}
}

func ExampleGenerateBlog() {
	fmt.Println(Must(gen.String(&markup.BlogGenTestStruct)))
	// Output:
	// <???>
}
