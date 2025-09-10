package executer

import (
	"fmt"
	"os"
	"shelly/app/syscallHelpers"
	"strings"
	"unsafe"
)

var newLineSlice = []byte{'\n'}

func handleType(args []string, outFd uintptr) {

	if len(args) != 1 {
		fmt.Println("No arg present when handling the type command")
	}

	arg := args[0]

	//We can then have a hashmap holding this
	switch arg {
	case "echo", "exit", "type", "pwd", "cd":
		byteArg := unsafe.Slice(unsafe.StringData(arg), len(arg))
		syscallHelpers.WriteWithSyscall(int(outFd), byteArg)

		byteResponse := " is a shell builtin\n"
		byteData := unsafe.Slice(unsafe.StringData(byteResponse), len(byteResponse))
		syscallHelpers.WriteWithSyscall(int(outFd), byteData)

		return
	}

	fullPath, _ := findExecutableBinaryInPath(arg)
	if fullPath != "" {
		//Simple way to write to either console or file

		// Preallocate with enough capacity to hold all parts
		// Heap allocated.
		msg := make([]byte, 0, len(arg)+len(" is ")+len(fullPath)+1)
		msg = append(msg, arg...)
		msg = append(msg, " is "...)
		msg = append(msg, fullPath...)
		msg = append(msg, '\n')

		syscallHelpers.WriteWithSyscall(int(outFd), msg)
		return
	}

	byteArgData := unsafe.Slice(unsafe.StringData(arg), len(arg))
	syscallHelpers.WriteWithSyscall(int(outFd), byteArgData)

	response := ": not found\n"
	byteResponse := unsafe.Slice(unsafe.StringData(response), len(response))
	syscallHelpers.WriteWithSyscall(int(outFd), byteResponse)
}

func handlePWD(outFd uintptr) {
	currentWorkingDirectory, _ := os.Getwd()
	byteCurrentWorkingDirectory := unsafe.Slice(unsafe.StringData(currentWorkingDirectory), len(currentWorkingDirectory))
	syscallHelpers.WriteWithSyscall(int(outFd), byteCurrentWorkingDirectory)
	syscallHelpers.WriteWithSyscall(int(outFd), newLineSlice)
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
