package main

import (
	"bufio"
	"fmt"
	"os"
	"shelly/app/executer"
	"shelly/app/parser"
	"shelly/app/parser/lexer"
	"shelly/app/parser/token"
)

func main() {

	for true {

		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		input, _ := bufio.NewReader(os.Stdin).ReadString('\n')

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
