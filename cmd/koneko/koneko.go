package main

import (
	"fmt"
	"os"

	"github.com/cvanloo/blog-go/markup"
)

func main() {
	fmt.Println("Hello, 子猫ちゃん")
	m := markup.New(
		markup.Source("<stdin>", os.Stdin),
		markup.OutDir("/tmp/koneko-out"),
	)
	if err := m.Run(); err != nil {
		fmt.Println(err)
	}
}
