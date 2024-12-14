package gen_test

import (
	"fmt"

	. "github.com/cvanloo/blog-go/assert"
	"github.com/cvanloo/blog-go/markup"
	"github.com/cvanloo/blog-go/markup/gen"
)

func ExampleGenerateBlog() {
	fmt.Println(Must(gen.String(&markup.BlogTestStruct)))
	// Output:
	// <???>
}
