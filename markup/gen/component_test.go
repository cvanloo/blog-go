package gen_test

import (
	"fmt"

	"github.com/cvanloo/blog-go/markup"
	"github.com/cvanloo/blog-go/markup/gen"
)

func must[T any](t T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("must: %s", err))
	}
	return t
}

func asRef[T any](t T) *T {
	return &t
}

func ExampleGenerateBlog() {
	fmt.Println(must(gen.String(markup.BlogTestStruct)))
	// Output:
	// <???>
}
