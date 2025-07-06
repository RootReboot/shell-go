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

	return token.Token{Type: token.TokenWord, Value: builder.String()}
}

func (l *Lexer) readEscapeCharacter() byte {
	//Right now we are in the escape character index
	//We want to go to the next one.
	//Two pos are jumped to also jump the character being escaped
	l.pos = l.pos + 2

	return l.input[l.pos-1]
}

// // readEscapeCharacter returns the result of processing an escape sequence.
// // It assumes the current position is at the backslash '\' and returns a valid escaped byte.
// // For unknown escapes, it returns the backslash and following character literally.
// func (l *Lexer) readEscapeCharacter() byte {
// 	if l.pos+1 >= len(l.input) {
// 		// Backslash at end of input — return it literally
// 		l.pos++
// 		return '\\'
// 	}

// 	escaped := l.input[l.pos+1]
// 	var result byte

// 	switch escaped {
// 	case 'n':
// 		result = '\n'
// 	case 't':
// 		result = '\t'
// 	case '\\':
// 		result = '\\'
// 	case '"':
// 		result = '"'
// 	case '\'':
// 		result = '\''
// 	default:
// 		// Unknown escape — treat \x as literal '\', 'x'
// 		// Caller can handle appending both if needed
// 		result = 0 // Signal that this is not a valid escape
// 	}

// 	if result != 0 {
// 		l.pos += 2
// 		return result
// 	}

// 	// Return the backslash literally and let the caller decide
// 	l.pos += 2
// 	// NOTE: Returning backslash and next char would need more logic,
// 	// so here we just return backslash and let the caller handle the second byte
// 	return '\\'
// }

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
