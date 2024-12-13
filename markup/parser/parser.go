package parser

import (
	"fmt"

	"github.com/cvanloo/blog-go/markup/lexer"
	"github.com/cvanloo/blog-go/markup/gen"
)

type (
	Token interface {
		Type() lexer.TokenType
		Text() string
		Location() string
	}
	ParserError struct {
		Token Token
		Inner error
	}
)

func (err ParserError) Error() string {
	return fmt.Sprintf("%s: %s", err.Token.Location(), err.Inner)
}

func Parse(tokens []Token) (*gen.Blog, error) {
}
