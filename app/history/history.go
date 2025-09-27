package history

/*
#cgo LDFLAGS: -lreadline
#include <stdlib.h>
#include <stdio.h>
#include <readline/readline.h>
#include <readline/history.h>
#include "../readline_helper.h"
*/
import "C"

import (
	"fmt"
	"os"
	"shelly/app/syscallHelpers"
	"strconv"
	"sync"
	"syscall"
	"unsafe"
)

// HistoryManager is a singleton controlling all history operations
type HistoryManager struct {
	histSize             int
	histFileSize         int
	fileStates           map[string]*fileState // per-file tracking
	defaultHistfilePath  string                // default HISTFILE path
	defaultHistfilePathC *C.char               // C string for default file
	countSessionHistory  int                   // counts the commands that were written in this section
	loaded               bool
	initOnce             sync.Once
}

type fileState struct {
	startLength  int // number of entries loaded from this file
	appendOffset int // number of entries already appended
}

var instance *HistoryManager
var once sync.Once

// GetHistoryManager returns the singleton HistoryManager instance.
func GetHistoryManager() *HistoryManager {
	once.Do(func() {
		instance = &HistoryManager{}
		instance.init()
	})
	return instance
}

func (h *HistoryManager) init() {
	// Bootstraqp the readline ihstory internal structures
	C.using_history()

	// Default values like in bash
	h.histSize = getEnvAsInt("HISTSIZE", 500)
	h.histFileSize = getEnvAsInt("HISTFILESIZE", 2000)

	h.fileStates = make(map[string]*fileState)

	// Determine history file path
	histfilePath, found := syscall.Getenv("HISTFILE")
	if !found {
		// Fallback to HOME + default bash history
		homePath, found := syscall.Getenv("HOME")
		if found {
			histfilePath = homePath + "/.bash_history"
		}
	}

	// Used to setup the command completion via TAB
	C.setup_completion()

	// Allocate C string for default file if present
	if histfilePath != "" {
		h.defaultHistfilePathC = C.CString(histfilePath)
		h.defaultHistfilePath = histfilePath
		// Load file if exists
		if _, err := os.Stat(histfilePath); err == nil {
			h.defaultHistfilePathC = C.CString(histfilePath)
			h.defaultHistfilePath = histfilePath
			C.read_history(h.defaultHistfilePathC)
			h.fileStates[histfilePath] = &fileState{
				startLength:  int(C.history_length),
				appendOffset: int(C.history_length),
			}
		} else {
			h.defaultHistfilePath = ""
			h.defaultHistfilePathC = nil
		}
	}
	h.loaded = true
}

// GetHistory returns the last `count` history entries as a slice of Go strings.
// If count < 0, it returns all entries.
func GetHistory(count int) []string {
	historyList := C.history_list()
	if historyList == nil {
		return nil
	}

	// Pointer arithmetic setup
	historyListStartPointer := uintptr(unsafe.Pointer(historyList))
	historyEntrySize := unsafe.Sizeof(*historyList)

	// Find total entries
	total := 0
	for i := 0; ; i++ {
		entryPointer := unsafe.Pointer(historyListStartPointer + uintptr(i)*historyEntrySize)
		entry := *(**C.HIST_ENTRY)(entryPointer)
		if entry == nil {
			total = i
			break
		}
	}

	// Determine starting index
	startIndex := 0
	if count >= 0 && count < total {
		startIndex = total - count
	}

	// Collect entries into a slice of Go strings
	result := make([]string, 0, total-startIndex)
	for i := startIndex; i < total; i++ {
		entryPointer := unsafe.Pointer(historyListStartPointer + uintptr(i)*historyEntrySize)
		entry := *(**C.HIST_ENTRY)(entryPointer)
		result = append(result, C.GoString(entry.line))
	}

	return result
}
func (h *HistoryManager) ReadLine(prompt string) string {
	cPrompt := C.CString(prompt)
	defer C.free(unsafe.Pointer(cPrompt))

	line := C.readline(cPrompt)
	if line == nil {
		return "" // EOF (Ctrl+D)
	}
	defer C.free(unsafe.Pointer(line))

	input := C.GoString(line)

	if input != "" {
		h.AddCommand(input) // automatically handle memory, HISTSIZE, HISTFILESIZE
	}

	return input
}

// AddCommand adds a new command to history with size enforcement.
func (h *HistoryManager) AddCommand(line string) {
	if !h.loaded || line == "" {
		return
	}

	cLine := C.CString(line)
	defer C.free(unsafe.Pointer(cLine))

	// Add to in-memory history
	C.add_history(cLine)
	h.countSessionHistory++

	// Enforce HISTSIZE for in-memory history
	if h.histSize > 0 && int(C.history_length) > h.histSize {
		C.remove_history(0)
	}

	// Do NOT write to file here (bash-like)
}

func (h *HistoryManager) WriteHistoryToFile(path string) {
	//If empty then it goes to the default history file
	if path == "" {
		//This means there is no file to append the data.
		if h.defaultHistfilePath == "" {
			return
		}

		path = h.defaultHistfilePath
	}

	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	// C.write_history creates the file if it does not exist
	// Try writing history to file
	if rc := C.write_history(cPath); rc != 0 {
		fmt.Printf("history -w: failed to write history to %s\n", path)
		return
	}

	// Update file state to reflect that everything is written
	fs, ok := h.fileStates[path]
	if !ok {
		fs = &fileState{}
		h.fileStates[path] = fs
	}
	fs.appendOffset = int(C.history_length)
	fs.startLength = fs.appendOffset
}

func (h *HistoryManager) AppendHistoryToFile(path string) {
	//If empty then it goes to the default history file
	if path == "" {
		//This means there is no file to append the data.
		if h.defaultHistfilePath == "" {
			return
		}

		path = h.defaultHistfilePath
	}

	fs, ok := h.fileStates[path]
	if !ok {
		historyLength := int(C.history_length)
		fs = &fileState{startLength: historyLength, appendOffset: 0}
		h.fileStates[path] = fs
	}

	newLines := h.countSessionHistory - fs.appendOffset
	if newLines > 0 {

		cPath := C.CString(path)
		defer C.free(unsafe.Pointer(cPath))

		// Append all new lines to the file; 0 means all new entries
		if rc := C.append_history(C.int(newLines), cPath); rc != 0 {
			fmt.Printf("history -a: failed to append history to %s\n", path)
			return
		}

		fs.appendOffset += newLines

		// If this is the default histfile, enforce HISTFILESIZE
		if path == h.defaultHistfilePath && h.histFileSize > 0 {
			// Use history_truncate_file from readline (if available)
			if rc := C.history_truncate_file(cPath, C.int(h.histFileSize)); rc != 0 {
				fmt.Printf("history -a: failed to truncate history in the file to %s\n", path)
				return
			}
		}
	}
}

func (h *HistoryManager) ReadHistory(path string) {

	err := syscallHelpers.FileExists(path)
	if err != nil {
		fmt.Printf("history -r: file not found: %s due to %v\n", path, err)
		return
	}

	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	if rc := C.read_history(cPath); rc != 0 {
		fmt.Printf("history -r: failed to read history from %s\n", path)
		return
	}
}

// Close frees resources associated with the history manager.
func (h *HistoryManager) Close() {
	if h.defaultHistfilePathC != nil {
		C.free(unsafe.Pointer(h.defaultHistfilePathC))
		h.defaultHistfilePathC = nil
	}
	h.loaded = false
}

func getEnvAsInt(name string, def int) int {
	val := os.Getenv(name)
	if val == "" {
		return def
	}
	if n, err := strconv.Atoi(val); err == nil {
		return n
	}
	return def
}
