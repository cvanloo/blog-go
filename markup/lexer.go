package markup

type (
	Lexer struct {
		Filename string
		Pos int
		Tokens []Token
		Errors []error
	}
	LexerError struct {
		Filename string
		Pos int
		Inner error
	}
	Token struct {
		Type TokenType
		Pos, Len int
	}
	TokenType int
)

func (err LexerError) Error() string {
	return fmt.Sprintf("%s:+%d: %s", err.Filename, err.Pos, err.Inner)
}
