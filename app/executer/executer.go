package executer

import (
	"fmt"
	"os"
	"shelly/app/parser/ast"
)

func RunPipeline(p ast.Pipeline) {
	for _, cmd := range p.Commands {
		runCommand(cmd)
	}
}

func runCommand(cmd ast.SimpleCommand) {
	if len(cmd.Args) == 0 {
		fmt.Println("Didn't find any data in the command")
	}

	cmdName := cmd.Args[0]
	args := cmd.Args[1:]

	switch cmdName {
	case "type":
		handleType(args)
	case "exit":
		successExit := handleExit(args)
		if successExit {
			os.Exit(0)
		}
	case "pwd":
		handlePWD()
	case "cd":
		handleCd(args)
	default:
		if err := runExecutable(cmd); err != nil {
			fmt.Println(err)
		}
	}
}
