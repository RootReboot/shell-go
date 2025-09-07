package ast

type Command interface{}

type SimpleCommand struct {
	Args      []string
	Redirects []Redirect // Ordered list
}

type Pipeline struct {
	Commands []SimpleCommand
	// To support redirects from a pipe operations
	Redirects []Redirect // Ordered list
}

type RedirectType int

const (
	RedirectStdout RedirectType = iota
	RedirectStdoutAppend
	RedirectStderr
	RedirectStderrAppend
)

type Redirect struct {
	Target string
	Type   RedirectType
}
