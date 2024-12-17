package main

import (
	"fmt"
	"os"
	"errors"
	"io"

	. "github.com/cvanloo/blog-go/assert"
	"github.com/cvanloo/blog-go/markup"
	"github.com/cvanloo/blog-go/markup/lexer"
)

func main() {
	inName := "-"
	inFile := os.Stdin
	inNameDisplay := "<stdin>"
	if len(os.Args) > 1 {
		inName = os.Args[1]
	}
	if inName != "-" {
		inNameDisplay = inName
		inFile = Must(os.Open(inName))
	}
	lx := lexer.New()
	lx.LexSource(inNameDisplay, string(Must(io.ReadAll(inFile))))
	if len(lx.Errors) > 0 {
		fmt.Println(errors.Join(lx.Errors...))
	}
	for token := range lx.Tokens() {
		fmt.Println(token)
	}
}

func main2() {
	fmt.Println("Hello, 子猫ちゃん")
	m := markup.New(
		markup.Source("<stdin>", os.Stdin),
		markup.OutDir("/tmp/koneko-out"),
	)
	if err := m.Run(); err != nil {
		fmt.Println(err)
	}
}
