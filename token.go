package strtotime

type tokenType int

const (
	typeString tokenType = iota + 1
)

type Token struct {
	val string
	typ tokenType
	pos int
}

// Tokenize takes a string and cuts it into tokens.
// * Special characters such as +
func Tokenize(s string) []Token {
	// TODO
	return nil
}
