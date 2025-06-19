package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {

	for true {

		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		input, _ := bufio.NewReader(os.Stdin).ReadString('\n')

		var fields = strings.Fields(input)

		//If nothing is inserted, then we show the prompt again.
		if len(fields) == 0 {
			continue
		}

		cmd := fields[0]

		//Type
		if cmd == "type" {

			if len(fields) != 2 {
				printCommandNotFound(cmd)
				continue
			}

			arg := fields[1]

			//We can then have a hashmap holding this
			switch arg {
			case "echo", "exit", "type":
				os.Stdout.WriteString(arg)
			default:
				fmt.Println(arg + ": command not found")
				continue
			}

			os.Stdout.WriteString(" is a shell builtin\n")
			continue
		}

		//Echo
		if cmd == "echo" {

			if len(fields) < 2 {
				printCommandNotFound(cmd)
				break
			}

			lastIndex := len(fields) - 1
			for i, word := range fields[1:] {
				if i > 0 && i < lastIndex {
					fmt.Print(" ")
				}
				fmt.Print(word)
			}
			fmt.Println()

			continue
		}

		//Exit
		if cmd == "exit" {
			if len(fields) != 2 || fields[1] != "0" {
				printCommandNotFound(cmd)
				continue
			}

			break
		}

		printCommandNotFound(cmd)
	}
}

func printCommandNotFound(cmd string) {
	fmt.Println(cmd + ": command not found")
}
