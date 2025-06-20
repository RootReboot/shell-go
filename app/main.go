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

		switch cmd {
		case "type":
			successType := HandleType(args)
			if successType {
				continue
			}
		case "echo":
			successEcho := HandleEcho(args)
			if successEcho {
				continue
			}
		case "exit":
			successExit := HandleExit(args)
			if successExit {
				os.Exit(0)
			}
		case "pwd":
			HandlePWD()
			continue
		case "cd":
			HandleCd(args)
			continue
		default:
			successRunOfExec := HandleExecutable(cmd, args)
			if successRunOfExec {
				continue
			}
		}

		printCommandNotFound(cmd)
	}
}

func HandleCd(args []string) {

	if len(args) != 1 {
		return
	}

	path := args[0]

	err := os.Chdir(path)
	if err != nil {
		fmt.Printf("cd: %s: No such file or directory", path)
	}
}

func HandlePWD() {
	currentWorkingDirectory, _ := os.Getwd()
	fmt.Println(currentWorkingDirectory)
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

func HandleType(args []string) bool {

	if len(args) != 1 {
		return false
	}

	arg := args[0]

	//We can then have a hashmap holding this
	switch arg {
	case "echo", "exit", "type", "pwd", "cd":
		os.Stdout.WriteString(arg)
		os.Stdout.WriteString(" is a shell builtin\n")
		return true
	}

	fullPath := FindExecutableInPath(arg)
	if fullPath != "" {
		fmt.Printf("%s is %s\n", arg, fullPath)
		return true
	}

	fmt.Println(arg + ": not found")

	return true
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

func HandleEcho(args []string) bool {
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

func HandleExit(args []string) bool {

	if len(args) != 1 || args[0] != "0" {
		return false
	}
	return true
}

func printCommandNotFound(cmd string) {
	fmt.Println(cmd + ": command not found")
}
