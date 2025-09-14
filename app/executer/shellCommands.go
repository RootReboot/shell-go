package executer

/*
#cgo LDFLAGS: -lreadline
#include <stdio.h>
#include <readline/history.h>
*/
import "C"

import (
	"fmt"
	"os"
	"shelly/app/syscallHelpers"
	"strconv"
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
	case "echo", "exit", "type", "pwd", "cd", "history":
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

// handleHistory prints the current Readline command history to stdout.
//
// It uses the GNU Readline C library's history API to access the in-memory
// history list. Each line entered by the user is stored as a HIST_ENTRY* in
// a NULL-terminated array.
//
// Parameters:
//   - args: currently unused, but could be extended to support options
//
// How it works:
// 1. Calls C.history_list() to get a pointer to the array of HIST_ENTRY*.
// 2. If the history list is empty (nil), the function returns immediately.
// 3. Converts the base pointer of the list to uintptr to allow pointer arithmetic.
// 4. Iterates through the array:
//   - Calculates the pointer to the i-th entry using pointer arithmetic.
//   - Dereferences the pointer to get the HIST_ENTRY*.
//   - Breaks the loop if the entry is nil (end of list).
//   - Converts the C string (entry.line) to a Go string and prints it with its index.
func handleHistory(args []string) {
	historyList := C.history_list()
	if historyList == nil {
		return // No history available
	}

	// Determine how many entries to print
	var count int = -1 // default: print all
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("history: %s: numeric argument required\n", args[0])
			return
		}

		if n >= 0 {
			count = n
		}
	}

	// Convert the start of the history list to uintptr for pointer arithmetic
	historyListStartPointer := uintptr(unsafe.Pointer(historyList))
	// Size of each HIST_ENTRY* element in the array
	historyEntrySize := unsafe.Sizeof(*historyList)

	total := 0
	for i := 0; ; i++ {
		// Calculate pointer to the i-th HIST_ENTRY*
		entryPointer := unsafe.Pointer(historyListStartPointer + uintptr(i)*historyEntrySize)
		// Dereference the pointer to get the actual HIST_ENTRY*
		entry := *(**C.HIST_ENTRY)(entryPointer)

		if entry == nil {
			total = i
			break // Reached the end of the history list
		}
	}

	startIndex := 0
	if count >= 0 && count < total {
		startIndex = total - count
	}

	for i := startIndex; i < total; i++ {

		// Calculate pointer to the i-th HIST_ENTRY*
		entryPointer := unsafe.Pointer(historyListStartPointer + uintptr(i)*historyEntrySize)
		// Dereference the pointer to get the actual HIST_ENTRY*
		entry := *(**C.HIST_ENTRY)(entryPointer)

		// Print the history index and the line content
		fmt.Println(i+1, C.GoString(entry.line))
	}

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
