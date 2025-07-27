package main

/*
#cgo LDFLAGS: -lreadline
#include <stdlib.h>
#include <stdio.h>
#include <readline/readline.h>
#include "readline_helper.h"
*/
import "C"

import (
	"shelly/app/executer"
	"shelly/app/parser"
	"shelly/app/parser/lexer"
	"shelly/app/parser/token"
	"unsafe"
)

func main() {

	C.setup_completion()

	for true {

		prompt := C.CString("$ ")
		line := C.readline(prompt)
		C.free(unsafe.Pointer(prompt))

		input := C.GoString(line)
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

// func HandleEcho(args []string) bool {
// 	if len(args) < 1 {
// 		return false
// 	}

// 	lastIndex := len(args) - 1
// 	for i, word := range args {
// 		fmt.Print(word)

// 		if i < lastIndex {
// 			fmt.Print(" ")
// 		}
// 	}
// 	fmt.Println()

// 	return true
// }
