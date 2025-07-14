package ast

type Command interface{}

type SimpleCommand struct {
	Args        []string
	RedirectOut *string
}

type Pipeline struct {
	Commands []SimpleCommand
	// To support redirects from a pipe operations
	RedirectOut *string
}
