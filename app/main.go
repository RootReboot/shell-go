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
