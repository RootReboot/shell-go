package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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
		args := fields[1:]

		//Type
		successType := HandleType(cmd, args)
		if successType {
			continue
		}

		//Echo
		successEcho := HandleEcho(cmd, args)
		if successEcho {
			continue
		}

		//Exit
		successExit := HandleExit(cmd, args)
		if successExit {
			continue
		}

		successRunOfExec := HandleExecutable(cmd, args)
		if successRunOfExec {
			continue
		}

		printCommandNotFound(cmd)
	}
}

func HandleExecutable(cmd string, args []string) bool {

	fullPath := FindExecutableInPath(cmd)
	if fullPath == "" {
		return false
	}

	execCMD := exec.Command(cmd, args...)
	execCMD.Stdout = os.Stdout
	execCMD.Stderr = os.Stderr
	execCMD.Run()

	return true
}

func HandleType(cmd string, args []string) bool {
	if cmd == "type" {

		if len(args) != 1 {
			return false
		}

		arg := args[0]

		//We can then have a hashmap holding this
		switch arg {
		case "echo", "exit", "type":
			os.Stdout.WriteString(arg)
			os.Stdout.WriteString(" is a shell builtin\n")
			return true
		}

		fullPath := FindExecutableInPath(arg)
		if fullPath != "" {
			fmt.Printf("%s is %s\n", arg, fullPath)
		}

		fmt.Println(arg + ": not found")

		return true
	}

	return false
}

func FindExecutableInPath(cmd string) string {
	pathEnvVar, envVarExists := os.LookupEnv("PATH")
	if envVarExists {

		pathsToCheck := strings.Split(pathEnvVar, ":")
		for _, path := range pathsToCheck {
			//Only supports single word argument
			fullPath := filepath.Join(path, cmd)

			fileInfo, err := os.Stat(fullPath)
			// Check if file exists and is executable
			if err == nil && fileInfo.Mode().Perm()&0111 != 0 {
				return fullPath
			}
		}
	}

	return ""
}

func HandleEcho(cmd string, args []string) bool {
	if cmd == "echo" {
		if len(args) < 1 {
			return false
		}

		lastIndex := len(args) - 1
		for i, word := range args {
			fmt.Print(word)

			if i < lastIndex {
				fmt.Print(" ")
			}
		}
		fmt.Println()

		return true
	}

	return false
}

func HandleExit(cmd string, args []string) bool {
	if cmd == "exit" {
		if len(args) != 1 || args[0] != "0" {
			return false
		}
		return true
	}
	return false
}

func printCommandNotFound(cmd string) {
	fmt.Println(cmd + ": command not found")
}
