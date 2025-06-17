package main

import (
	"bufio"
	"fmt"
	"os"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

func main() {

	for true {

		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		command, _ := bufio.NewReader(os.Stdin).ReadString('\n')

		//Type
		if len(command) > 4 && command[:5] == "type " {

			commandAsked := command[5:]
			//Having the string literals li
			switch commandAsked {
			case "echo\n":
				fmt.Print("echo")
			case "exit\n":
				fmt.Print("exit")
			case "type\n":
				os.Stdout.WriteString("type")
			default:
				fmt.Println(command[:len(command)-1] + ": command not found")
				continue
			}

			os.Stdout.WriteString(" is a shell builtin\n")
			continue
		}

		//Echo
		if len(command) > 5 && command[:5] == "echo " {
			fmt.Print(command[5:])
			continue
		}

		//Exit
		if command == "exit 0\n" {
			break
		}

		fmt.Println(command[:len(command)-1] + ": command not found")

	}
}
