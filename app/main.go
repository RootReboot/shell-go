package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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
				os.Stdout.WriteString(" is a shell builtin\n")
				continue
			}

			pathEnvVar, envVarExists := os.LookupEnv("PATH")
			if envVarExists {

				pathsToCheck := strings.Split(pathEnvVar, ":")
				commandIsPresent := false
				for _, path := range pathsToCheck {
					//Only supports single word argument
					fullPath := filepath.Join(path, arg)

					_, err := os.Stat(fullPath)
					if err == nil {
						fmt.Printf("%s is %s\n", arg, fullPath)
						commandIsPresent = true
						break
					}
				}

				if commandIsPresent {
					continue
				}
			}

			fmt.Println(arg + ": not found")

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
