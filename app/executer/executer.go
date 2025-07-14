package executer

import (
	"fmt"
	"io"
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

	outFile := os.Stdout
	if cmd.RedirectOut != nil {
		outFile, err := os.Create(*cmd.RedirectOut)
		if err != nil {
			fmt.Printf("Failed to open file for redirection: %v\n", err)
			return
		}
		defer outFile.Close()
	}

	var out io.Writer = outFile

	cmdName := cmd.Args[0]
	args := cmd.Args[1:]

	switch cmdName {
	case "type":
		handleType(args, out)
	case "exit":
		successExit := handleExit(args)
		if successExit {
			os.Exit(0)
		}
	case "pwd":
		handlePWD(out)
	case "cd":
		handleCd(args)
	default:
		if err := runExecutable(cmd, outFile); err != nil {
			fmt.Println(err)
		}
	}
}
