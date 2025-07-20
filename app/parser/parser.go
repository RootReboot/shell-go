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

	//TODO: This needs to be fixed
	for p.match(token.TokenPipe) {
		cmd, err := p.parseSimpleCommand()
		if err != nil {
			return pipeline, err
		}

		pipeline.Commands = append(pipeline.Commands, cmd)
	}

	if p.match(token.TokenRedirectOut) {
		p.pos++

		if !p.match(token.TokenWord) {
			return pipeline, fmt.Errorf("expected file name after '>")
		}

		pipeline.RedirectOut = &p.tokens[p.pos].Value
		p.pos++
	}

	if p.match(token.TokenRedirectErr) {
		// Advance the '2>'
		p.pos++

		if !p.match(token.TokenWord) {
			return pipeline, fmt.Errorf("expected file name after '2>")
		}

		pipeline.RedirectErr = &p.tokens[p.pos].Value
		p.pos++
	}

	return pipeline, nil
}

func (p *Parser) parseSimpleCommand() (ast.SimpleCommand, error) {
	var cmd ast.SimpleCommand

	if !p.match(token.TokenWord) {
		return cmd, fmt.Errorf("expected command")
	}

	//First word in the command
	cmd.Args = append(cmd.Args, p.tokens[p.pos].Value)
	p.pos++

	// Consume remaining arguments (also TokenWord)
	for p.match(token.TokenWord) {
		cmd.Args = append(cmd.Args, p.tokens[p.pos].Value)
		p.pos++
	}

	if p.match(token.TokenRedirectOut) {
		// Advance the '>' or '1>'
		p.pos++

		if !p.match(token.TokenWord) {
			return cmd, fmt.Errorf("expected file name after '>")
		}

		cmd.RedirectOut = &p.tokens[p.pos].Value

		//Move to the next pos
		p.pos++
	}

	if p.match(token.TokenRedirectErr) {
		// Advance the '2>'
		p.pos++

		if !p.match(token.TokenWord) {
			return cmd, fmt.Errorf("expected file name after '>")
		}

		cmd.RedirectErr = &p.tokens[p.pos].Value

		//Move to the next pos
		p.pos++
	}

	if p.match(token.TokenAppendRedirectOut) {
		// Advance the '>>' or '1>>'
		p.pos++

		if !p.match(token.TokenWord) {
			return cmd, fmt.Errorf("expected file name after '>")
		}

		cmd.AppendRedirectOut = &p.tokens[p.pos].Value

		//Move to the next pos
		p.pos++
	}

	if p.match(token.TokenAppendRedirectErr) {
		// Advance the '2>>'
		p.pos++

		if !p.match(token.TokenWord) {
			return cmd, fmt.Errorf("expected file name after '>")
		}

		cmd.AppendRedirectErr = &p.tokens[p.pos].Value

		//Move to the next pos
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
