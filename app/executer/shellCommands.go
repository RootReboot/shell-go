package executer

import (
	"fmt"
	"os"
	"strings"
)

func handleType(args []string) {

	if len(args) != 1 {
		fmt.Println("No arg present when handling the type command")
	}

	arg := args[0]

	//We can then have a hashmap holding this
	switch arg {
	case "echo", "exit", "type", "pwd", "cd":
		os.Stdout.WriteString(arg)
		os.Stdout.WriteString(" is a shell builtin\n")
		return
	}

	fullPath, _ := findExecutableBinaryInPath(arg)
	if fullPath != "" {
		fmt.Printf("%s is %s\n", arg, fullPath)
		return
	}

	fmt.Println(arg + ": not found")
}

func handlePWD() {
	currentWorkingDirectory, _ := os.Getwd()
	fmt.Println(currentWorkingDirectory)
}

func handleCd(args []string) {

	if len(args) != 1 {
		return
	}

	path := args[0]

	path = substituteHomeDirectoryCharacter(path)

	err := os.Chdir(path)
	if err != nil {
		fmt.Printf("cd: %s: No such file or directory\n", path)
	}
}

func handleExit(args []string) bool {

	if len(args) != 1 || args[0] != "0" {
		return false
	}
	return true
}

func substituteHomeDirectoryCharacter(path string) string {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}

	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home + path[1:] // skip the "~"
	}

	return path
}
