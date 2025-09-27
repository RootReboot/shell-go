package main

/*
#cgo LDFLAGS: -lreadline
#include <stdlib.h>
#include <stdio.h>
#include <readline/readline.h>
#include <readline/history.h>
#include "readline_helper.h"
*/
import "C"

import (
	"os"
	"os/signal"
	"shelly/app/executer"
	"shelly/app/history"
	"shelly/app/parser"
	"shelly/app/parser/ast"
	"shelly/app/parser/lexer"
	"shelly/app/parser/token"
	"syscall"
)

func main() {

	sigs := make(chan os.Signal, 1)
	// Notify the `sigs` channel when the process receives an interrupt or termination signal:
	// - SIGINT  : Interrupt signal (usually from Ctrl+C)
	// - SIGTERM : Termination signal (polite request to stop, e.g. from `kill`)
	// - SIGHUP  : Hangup signal (sent when terminal closes or service manager restarts the process)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		//Blocks channel until new message
		<-sigs
		//"" defaults to the default history file.
		history.GetHistoryManager().AppendHistoryToFile("")
		os.Exit(0)
	}()

	// Check if we are in "run single command mode"
	// This is used when a builtin command like cd is used in a pipeline
	if len(os.Args) > 1 && os.Args[0] == "--run-builtin" {
		// os.Args[2:] is the command and its args
		cmd := ast.SimpleCommand{
			Args: os.Args[1:],
		}
		// Run the command and exit immediately
		if err := executer.RunSingleCommand(cmd); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	historyManager := history.GetHistoryManager()

	for true {

		prompt := "$ "

		input := historyManager.ReadLine(prompt)

		var lex = lexer.NewLexer(input)
		tokens := []token.Token{}
		for {
			tok := lex.NextToken()
			tokens = append(tokens, tok)
			if tok.Type == token.TokenEOF {
				history.GetHistoryManager().AppendHistoryToFile("")
				break
			}
		}

		var cmdParser = parser.NewParser(tokens)

		astTree, _ := cmdParser.Parse()

		executer.RunPipeline(astTree)
	}
}
