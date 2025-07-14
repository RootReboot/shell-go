// Package lexer provides functionality to tokenize input strings
// for parsing shell-like syntax (e.g., command pipelines).
package lexer

import (
	"shelly/app/parser/token"
	"strings"
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

	if l.input[l.pos] == '|' {
		l.pos++
		return token.Token{Type: token.TokenPipe, Value: "|"}
	}

	var builder strings.Builder

	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		if isWhiteSpace(ch) || ch == '|' {
			break
		}

		switch ch {
		case '"', '\'':
			quoted := l.readQuoted()
			builder.WriteString(quoted)
		case '\\':
			escapedChar := l.readEscapeCharacter()
			builder.WriteByte(escapedChar)
		default:
			// Regular unquoted character
			builder.WriteByte(ch)
			l.pos++
		}
	}

	tokenValue := builder.String()

	//cheaper to check the length in most cases then doing string comparison
	if len(tokenValue) == 1 || len(tokenValue) == 2 {
		if tokenValue == "1>" || tokenValue == ">" {
			return token.Token{Type: token.TokenRedirectOut, Value: tokenValue}
		}
	}

	if len(tokenValue) == 2 {
		if tokenValue == "1>" {
			return token.Token{Type: token.TokenRedirectOut, Value: tokenValue}
		}

		if tokenValue == "2>" {
			return token.Token{Type: token.TokenRedirectErr, Value: tokenValue}
		}
	}

	return token.Token{Type: token.TokenWord, Value: tokenValue}
}

func (l *Lexer) readEscapeCharacter() byte {
	//Right now we are in the escape character index
	//We want to go to the next one.
	//Two pos are jumped to also jump the character being escaped
	l.pos = l.pos + 2

	return l.input[l.pos-1]
}

func (l *Lexer) readQuoted() string {
	quote := l.input[l.pos]
	//skip opening quote
	l.pos++

	start := l.pos
	var builder strings.Builder
	for l.pos < len(l.input) && l.input[l.pos] != quote {

		if l.input[l.pos] == '\\' {
			builder.WriteString(l.input[start:l.pos]) // Flush before backslash

			if l.pos+1 < len(l.input) {
				escaped := l.input[l.pos+1]

				switch escaped {
				case quote, '\\', '$', '`':
					// Supported escape sequences
					switch escaped {
					default:
						builder.WriteByte(escaped)
					}
					l.pos += 2
				default:
					// Not a valid escape — keep backslash as-is
					builder.WriteByte('\\')
					builder.WriteByte(escaped)
					l.pos += 2
				}
			} else {
				// Lone backslash at end — treat as literal
				builder.WriteByte('\\')
				l.pos++
			}

			start = l.pos
			continue
		}

		l.pos++
	}

	// If closing quote was not found, return the rest as-is (unterminated string)
	if l.pos >= len(l.input) {
		return l.input[start-1:] // Include opening quote
	}

	// Append remaining content before closing quote
	builder.WriteString(l.input[start:l.pos])
	l.pos++ // Skip closing quote

	return builder.String()

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
