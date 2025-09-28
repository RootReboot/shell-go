package executer

import (
	"fmt"
	"io"
	"os"
	"shelly/app/history"
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
//   - args: Support options
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

	// historySoFar := history.GetHistory(count)

	// for _, line := range historySoFar {
	// 	// Print the history index and the line content
	// 	fmt.Println(line.Index, line.Line)
	// }

	history.ForEachHistory(count, PrintHistory)
}

func PrintHistory(idx int, line unsafe.Pointer) {
	fmt.Fprintf(os.Stdout, "%d ", idx)
	PrintCStr(os.Stdout, line)
	fmt.Fprintln(os.Stdout)
}

func PrintCStr(w io.Writer, cstr unsafe.Pointer) {
	ptr := uintptr(cstr)
	length := 0
	//C doesn't buffer the string size, so we need to figure out the size. C.strlen could also be used
	for *(*byte)(unsafe.Pointer(ptr + uintptr(length))) != 0 {
		length++
	}

	// Convert to slice pointing directly to C memory (no allocation)
	b := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), length)
	w.Write(b)
}

// loadHistoryFromFile loads history entries from a file into the readline history.
//
// Usage: history -r [filename]
// If no filename is provided, defaults to ~/.history or another configured file.
func loadHistoryFromFile(args []string) {
	var filename string
	if len(args) > 1 {
		filename = args[1]
	}

	history.GetHistoryManager().ReadHistory(filename)
}

// loadHistoryToFile saves the current readline history to a file.
//
// Usage: history -w [filename]
// If no filename is provided, nothing happens
func loadHistoryToFile(args []string) {
	var filename string
	if len(args) > 1 {
		filename = args[1]
	}

	history.GetHistoryManager().WriteHistoryToFile(filename)
}

// appendHistoryToFile appends new history entries to a file.a
//
// Usage: history -a [filename]
// If no filename is provided, defaults to ~/.history
func appendHistoryToFile(args []string) {
	var filename string
	if len(args) > 1 {
		filename = args[1]
	}

	history.GetHistoryManager().AppendHistoryToFile(filename)
}
