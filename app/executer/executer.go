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
	errFile := os.Stderr

	if cmd.RedirectOut != nil {
		var err error
		//The := created a new var. So if I used the := it wouldn't override the outfile outside the if scope
		outFile, err = os.Create(*cmd.RedirectOut)
		if err != nil {
			fmt.Printf("Failed to open file for redirection out: %v\n", err)
			return
		}
		defer outFile.Close()
	} else if cmd.RedirectErr != nil {
		var err error
		//The := created a new var. So if I used the := it wouldn't override the outfile outside the if scope
		errFile, err = os.Create(*cmd.RedirectErr)
		if err != nil {
			fmt.Printf("Failed to open file for redirection rr: %v\n", err)
			return
		}
		defer errFile.Close()
	}

	if cmd.AppendRedirectOut != nil {
		var err error
		//The := created a new var. So if I used the := it wouldn't override the outfile outside the if scope

		//To check what the flags do -> https://pubs.opengroup.org/onlinepubs/7908799/xsh/open.html
		outFile, err = os.OpenFile(*cmd.AppendRedirectOut, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			fmt.Printf("Failed to open file for append redirection out: %v\n", err)
			return
		}
		defer outFile.Close()
	} else if cmd.AppendRedirectErr != nil {
		var err error
		//The := created a new var. So if I used the := it wouldn't override the outfile outside the if scope

		//To check what the flags do -> https://pubs.opengroup.org/onlinepubs/7908799/xsh/open.html
		errFile, err = os.OpenFile(*cmd.AppendRedirectErr, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			fmt.Printf("Failed to open file for append redirection err: %v\n", err)
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
		if err := runExecutable(cmd, outFile, errFile); err != nil {
			fmt.Println(err)
		}
	}
}
