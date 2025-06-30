package parser

import (
	"fmt"
	"shelly/app/parser/ast"
	"shelly/app/parser/token"
)

type Parser struct {
	tokens []token.Token
	pos    int
}

func NewParser(tokens []token.Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) Parse() (ast.Pipeline, error) {

	var pipeline ast.Pipeline

	cmd, err := p.parseSimpleCommand()

	if err != nil {
		return pipeline, err
	}

	pipeline.Commands = append(pipeline.Commands, cmd)

	for p.match(token.TokenPipe) {
		cmd, err := p.parseSimpleCommand()
		if err != nil {
			return pipeline, err
		}

		pipeline.Commands = append(pipeline.Commands, cmd)
	}

	return pipeline, nil
}

func (p *Parser) parseSimpleCommand() (ast.SimpleCommand, error) {
	var cmd ast.SimpleCommand

	if !p.match(token.TokenWord) {
		return cmd, fmt.Errorf("expected command")
	}

	cmd.Args = append(cmd.Args, p.tokens[p.pos].Value)
	p.pos++

	// Consume remaining arguments (also TokenWord)
	for p.match(token.TokenWord) {
		cmd.Args = append(cmd.Args, p.tokens[p.pos].Value)
		p.pos++
	}

	return cmd, nil
}

func (p *Parser) match(tt token.TokenType) bool {
	return p.peek().Type == tt
}

func (p *Parser) peek() token.Token {
	if p.pos >= len(p.tokens) {
		return token.Token{Type: token.TokenEOF}
	}

	return p.tokens[p.pos]
}
