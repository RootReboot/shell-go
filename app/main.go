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
	"shelly/app/executer"
	"shelly/app/parser"
	"shelly/app/parser/ast"
	"shelly/app/parser/lexer"
	"shelly/app/parser/token"
	"unsafe"
)

func main() {

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

	C.setup_completion()

	for true {

		prompt := C.CString("$ ")
		line := C.readline(prompt)
		C.free(unsafe.Pointer(prompt))

		input := C.GoString(line)
		C.add_history(line)
		C.free(unsafe.Pointer(line))

		// fmt.Fprint(os.Stdout, "$ ")

		// // Wait for user input
		// input, _ := bufio.NewReader(os.Stdin).ReadString('\n')

		//Future implementation
		// if input != "" {
		// 	C.add_history(C.CString(input))
		// }

		var lex = lexer.NewLexer(input)
		tokens := []token.Token{}
		for {
			tok := lex.NextToken()
			tokens = append(tokens, tok)
			if tok.Type == token.TokenEOF {
				break
			}
		}

		var cmdParser = parser.NewParser(tokens)

		astTree, _ := cmdParser.Parse()

		executer.RunPipeline(astTree)
	}
}
