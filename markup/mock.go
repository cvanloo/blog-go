package markup

import (
	//"time"

	"github.com/cvanloo/blog-go/markup/lexer"
	"github.com/cvanloo/blog-go/markup/gen"
	//. "github.com/cvanloo/blog-go/assert"
)

type MockTokens []lexer.Token

func (m MockTokens) Tokens() func(func(lexer.Token) bool) {
	return func(yield func(lexer.Token) bool) {
		for _, t := range m {
			if !yield(t) {
				return
			}
		}
	}
}

func M(t lexer.TokenType, s string) lexer.Token {
	return lexer.Token{
		Type: t,
		Text: s,
	}
}

const BlogTestSource = ``

var LexerTestTokens = MockTokens{}

var BlogTestStruct = gen.Blog{}
