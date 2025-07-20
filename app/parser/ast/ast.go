package ast

type Command interface{}

type SimpleCommand struct {
	Args              []string
	AppendRedirectOut *string
	AppendRedirectErr *string
	RedirectOut       *string
	RedirectErr       *string
}

type Pipeline struct {
	Commands []SimpleCommand
	// To support redirects from a pipe operations
	AppendRedirectOut *string
	AppendRedirectErr *string
	RedirectOut       *string
	RedirectErr       *string
}
