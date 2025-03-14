package strtotime

import (
	"strings"
	"unicode"
)

type tokenType int

const (
	TypeString tokenType = iota + 1
	TypeNumber
	TypeOperator
	TypeWhitespace
	TypePunctuation
)

// Token represents a token extracted from the input string
type Token struct {
	Val string
	Typ tokenType
	Pos int
}

// Tokenize takes a string and cuts it into tokens.
func Tokenize(s string) []Token {
	var tokens []Token
	var currentToken strings.Builder
	var currentType tokenType
	var pos int

	for i, r := range s {
		var newType tokenType

		switch {
		case unicode.IsSpace(r):
			newType = TypeWhitespace
		case unicode.IsDigit(r):
			newType = TypeNumber
		case isOperator(r):
			newType = TypeOperator
		case isPunctuation(r):
			newType = TypePunctuation
		default:
			newType = TypeString
		}

		// If we're starting a new token or changing token type
		if i == 0 || newType != currentType {
			// Add the previous token if it exists
			if currentToken.Len() > 0 {
				tokens = append(tokens, Token{
					Val: currentToken.String(),
					Typ: currentType,
					Pos: pos,
				})
				currentToken.Reset()
			}

			currentType = newType
			pos = i
		}

		currentToken.WriteRune(r)
	}

	// Add the last token if it exists
	if currentToken.Len() > 0 {
		tokens = append(tokens, Token{
			Val: currentToken.String(),
			Typ: currentType,
			Pos: pos,
		})
	}

	return tokens
}

// isOperator checks if a rune is an operator
func isOperator(r rune) bool {
	return r == '+' || r == '-' || r == ':' || r == '/' || r == '.'
}

// isPunctuation checks if a rune is punctuation
func isPunctuation(r rune) bool {
	return r == ',' || r == ';' || r == '(' || r == ')' || r == '[' || r == ']'
}
