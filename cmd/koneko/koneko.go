// Example invocation:
// koneko -source hello_world.md -out /tmp/koneko
//
// koneko -source hello_world.md,goodbye_moon.md -out /tmp/koneko
//
// koneko -source hello_world.md -source goodbye_moon.md -out /tmp/koneko
package main

import (
	"fmt"
	"flag"
	"os"
	"strings"
	"log"

	//. "github.com/cvanloo/blog-go/assert"
	"github.com/cvanloo/blog-go/markup"
	//"github.com/cvanloo/blog-go/markup/lexer"
)

type ArrayFlag []string

func (af *ArrayFlag) String() string {
	return fmt.Sprintf("%v", *af)
}

func (af *ArrayFlag) Set(value string) error {
	values := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' '
	})
	*af = append(*af, values...)
	return nil
}

var (
	source ArrayFlag
	out = flag.String("out", ".", "Directory to write static sites to.")
)

func init() {
	flag.Var(&source, "source", "Input files. If given a directory, it will be processed recursively. A hyphen (the default) will read from stdin.")
	flag.Parse()
}

func main() {
	os.Exit(app())
}

func app() int {
	fmt.Println("Hello, 子猫ちゃん")
	fi, err := os.Stat(*out)
	if err != nil {
		log.Println(err)
		return -1
	}
	if !fi.IsDir() {
		log.Printf("%s is not a directory", *out)
		return -1
	}
	m := markup.New(
		markup.SourcePaths(source),
		markup.OutDir(*out),
	)
	if err := m.Run(); err != nil {
		fmt.Println(err)
	}

	return 0
}
