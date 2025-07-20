package main

/*
#cgo LDFLAGS: -lreadline
#include <stdlib.h>
#include <string.h>
#include <readline/readline.h>
#include <readline/history.h>

// Custom generator for a fixed set of completions
char* commands[] = {"help", "exit", "run", "status", "echo", "type", NULL};

// Generator function called repeatedly by readline
char* my_generator(const char* text, int state) {
    static int list_index, len;
    char* name;

    if (!state) {
        list_index = 0;
        len = strlen(text);
    }

    while ((name = commands[list_index++])) {
        if (strncmp(name, text, len) == 0) {
            return strdup(name);
        }
    }

    return NULL;
}

// Called by readline when tab is pressed
char** my_completion(const char* text, int start, int end) {
    return rl_completion_matches(text, my_generator);
}

// Set the completion function
void setup_completion() {
    rl_attempted_completion_function = my_completion;
}
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

		// fmt.Fprint(os.Stdout, "$ ")

		// // Wait for user input
		// input, _ := bufio.NewReader(os.Stdin).ReadString('\n')


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
