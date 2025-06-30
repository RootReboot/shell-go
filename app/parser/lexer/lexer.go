// Package lexer provides functionality to tokenize input strings
// for parsing shell-like syntax (e.g., command pipelines).
package lexer

import (
	"shelly/app/parser/token"
)

// Lexer represents a lexical analyzer that tokenizes a shell input string.
type Lexer struct {
	input string // the input string to tokenize
	pos   int    // current reading position in the input
}

// NewLexer returns a new instance of Lexer initialized with the given input.
func NewLexer(input string) *Lexer {
	return &Lexer{input: input}
}

// NextToken returns the next token from the input.
// It skips any leading whitespace and returns tokens such as:
// - token.TokenEOF when input is exhausted
// - token.TokenPipe for pipe characters '|'
// - token.TokenWord for literal words (non-whitespace, non-special chars)
func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()

	// End of input: return EOF token
	if l.pos >= len(l.input) {
		return token.Token{Type: token.TokenEOF}
	}

	switch l.input[l.pos] {
	case '|':
		l.pos++
		return token.Token{Type: token.TokenPipe, Value: "|"}
	case '"':
		return l.readQuoted('"')
	case '\'':
		return l.readQuoted('\'')
	}

	// Parse a word token
	start := l.pos
	for l.pos < len(l.input) && !isWhiteSpace(l.input[l.pos]) && l.input[l.pos] != '|' {
		//Special cases are dealt above
		if l.input[l.pos] == '|' || l.input[l.pos] == '"' || l.input[l.pos] == '\'' {
			break
		}

		l.pos++
	}
	return token.Token{Type: token.TokenWord, Value: l.input[start:l.pos]}
}

func (l *Lexer) readQuoted(quote byte) token.Token {
	//skip opening quote
	l.pos++

	start := l.pos

	for l.pos < len(l.input) && l.input[l.pos] != quote {
		l.pos++
	}

	//Handles incomplete cases(e.g: "hello how are you)
	if l.pos >= len(l.input) {
		//Since it is incomplete we return the quote in the token
		val := l.input[start-1:]
		return token.Token{Type: token.TokenWord, Value: val}
	}

	val := l.input[start:l.pos]
	//Skip closing quote
	l.pos++

	return token.Token{Type: token.TokenWord, Value: val}

}

// skipWhitespace advances the position pointer past any whitespace characters.
func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && isWhiteSpace(l.input[l.pos]) {
		l.pos++
	}
}

// isWhiteSpace returns true if the character is a whitespace character (space, tab, newline).
func isWhiteSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}
