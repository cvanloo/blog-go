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
	"errors"
	"io"
	"strings"
	"log"

	. "github.com/cvanloo/blog-go/assert"
	"github.com/cvanloo/blog-go/markup"
	"github.com/cvanloo/blog-go/markup/lexer"
)

func main2() {
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

type ArrayFlag []string

func (af *ArrayFlag) String() string {
	return fmt.Sprintf("%v", *af)
}

func (af *ArrayFlag) Set(value string) error {
	values := strings.Split(value, ",")
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
	sourceNames, sourceFDs, err := openSources(source)
	if err != nil {
		log.Println(err)
		return -1
	}
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
		markup.FileSources(sourceNames, sourceFDs),
		markup.OutDir(*out),
	)
	if err := m.Run(); err != nil {
		fmt.Println(err)
	}

	return 0
}

func openSources(paths []string) (names []string, fds []*os.File, err error) {
	if len(paths) == 1 && paths[0] == "-" {
		return []string{"<stdin>"}, []*os.File{os.Stdin}, nil
	}
	for _, path := range paths {
		fd, openErr := os.Open(path)
		if openErr != nil {
			err = errors.Join(err, openErr)
			continue
		}
		names = append(names, path)
		fds = append(fds, fd)
	}
	return names, fds, err
}
