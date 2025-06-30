package ast

type Command interface{}

type SimpleCommand struct {
	Args []string
}

type Pipeline struct {
	Commands []SimpleCommand
}
