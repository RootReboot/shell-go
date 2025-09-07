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

	// Pre size the array with 4. To avoid copies and creation of new arrays underneath the slice
	pipeline.Redirects = make([]ast.Redirect, 0, 4)

	if err != nil {
		return pipeline, err
	}

	pipeline.Commands = append(pipeline.Commands, cmd)

	for p.match(token.TokenPipe) {
		p.pos++
		cmd, err := p.parseSimpleCommand()
		if err != nil {
			return pipeline, err
		}

		pipeline.Commands = append(pipeline.Commands, cmd)
	}

	if err := AddRedirectInfo(p, &pipeline.Redirects); err != nil {
		return pipeline, err
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

	// Pre size the array with 4. To avoid copies and creation of new arrays underneath the slice
	cmd.Redirects = make([]ast.Redirect, 0, 4)

	if err := AddRedirectInfo(p, &cmd.Redirects); err != nil {
		return cmd, err
	}

	return cmd, nil
}

func AddRedirectInfo(p *Parser, redirects *[]ast.Redirect) error {
	if p.match(token.TokenRedirectOut) {
		p.pos++

		if !p.match(token.TokenWord) {
			return fmt.Errorf("expected file name after >")
		}

		stdoutRedirect := ast.Redirect{Target: p.tokens[p.pos].Value, Type: ast.RedirectStdout}
		*redirects = append(*redirects, stdoutRedirect)
		p.pos++
	}

	if p.match(token.TokenRedirectErr) {
		// Advance the '2>'
		p.pos++

		if !p.match(token.TokenWord) {
			return fmt.Errorf("expected file name after 2>")
		}

		stderrRedirect := ast.Redirect{Target: p.tokens[p.pos].Value, Type: ast.RedirectStderr}
		*redirects = append(*redirects, stderrRedirect)
		p.pos++
	}

	if p.match(token.TokenAppendRedirectOut) {
		p.pos++

		if !p.match(token.TokenWord) {
			return fmt.Errorf("expected file name after >>")
		}

		stdoutAppendRedirect := ast.Redirect{Target: p.tokens[p.pos].Value, Type: ast.RedirectStdoutAppend}
		*redirects = append(*redirects, stdoutAppendRedirect)
		p.pos++
	}

	if p.match(token.TokenAppendRedirectErr) {
		p.pos++

		if !p.match(token.TokenWord) {
			return fmt.Errorf("expected file name after >>")
		}

		stderrAppendRedirect := ast.Redirect{Target: p.tokens[p.pos].Value, Type: ast.RedirectStderrAppend}
		*redirects = append(*redirects, stderrAppendRedirect)
		p.pos++
	}

	return nil
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
