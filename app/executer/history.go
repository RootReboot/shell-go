package executer

/*
#cgo LDFLAGS: -lreadline
#include <stdio.h>
#include <stdlib.h>
#include <readline/history.h>
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"shelly/app/syscallHelpers"
	"strconv"
	"unsafe"
)

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

	if len(args) > 0 {
		switch args[0] {
		case "-r":
			loadHistoryFromFile(args)
			return
		case "-w":
			loadHistoryToFile(args)
			return
		case "-a":
			appendHistoryToFile(args)
			return
		}

	}

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

// loadHistoryFromFile loads history entries from a file into the readline history.
//
// Usage: history -r [filename]
// If no filename is provided, defaults to ~/.history or another configured file.
func loadHistoryFromFile(args []string) {
	var filename string

	if len(args) > 1 {
		filename = args[1]
	} else {
		// Default to ~/.history
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("history -r: could not determine home directory")
			return
		}
		filename = filepath.Join(home, ".history")
	}

	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	err := syscallHelpers.FileExists(filename)
	if err != nil {
		fmt.Printf("history -r: file not found: %s due to %v\n", filename, err)
		return
	}

	if rc := C.read_history(cFilename); rc != 0 {
		fmt.Printf("history -r: failed to read history from %s\n", filename)
		return
	}
}

// loadHistoryToFile saves the current readline history to a file.
//
// Usage: history -w [filename]
// If no filename is provided, nothing happens
func loadHistoryToFile(args []string) {
	var filename string

	if len(args) > 1 {
		filename = args[1]
	} else {
		// Default to ~/.history
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("history -w: could not determine home directory")
			return
		}
		filename = filepath.Join(home, ".history")
	}

	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	// C.write_history creates the file if it does not exist
	// Try writing history to file
	if rc := C.write_history(cFilename); rc != 0 {
		fmt.Printf("history -w: failed to write history to %s\n", filename)
		return
	}
}

// appendHistoryToFile appends new history entries to a file.a
//
// Usage: history -a [filename]
// If no filename is provided, defaults to ~/.history
func appendHistoryToFile(args []string) {
	var filename string
	if len(args) > 1 {
		filename = args[1]
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("history -a: could not determine home directory")
			return
		}
		filename = filepath.Join(home, ".history")
	}

	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	// Append all new lines to the file; 0 means all new entries
	if rc := C.append_history(0, cFilename); rc != 0 {
		fmt.Printf("history -a: failed to append history to %s\n", filename)
		return
	}
}
