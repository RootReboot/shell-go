package executer

import (
	"fmt"
	"os"
	"shelly/app/nullable"
	"shelly/app/parser/ast"
	"syscall"
)

func RunPipeline(p ast.Pipeline) {

	// default stdout/stderr for the whole pipeline
	stdoutFdPipe := os.Stdout.Fd()
	stderrFdPipe := os.Stderr.Fd()

	redirectStdoutNullable, redirectStderrNullable, err := setupRedirectsFd(p.Redirects)

	if err != nil {
		return
	}

	if redirectStdout, hasValue := redirectStdoutNullable.Get(); hasValue {
		stdoutFdPipe = uintptr(redirectStdout)

		defer syscall.Close(redirectStdout)
	}

	if redirectStderr, hasValue := redirectStderrNullable.Get(); hasValue {
		stderrFdPipe = uintptr(redirectStderr)

		defer syscall.Close(redirectStderr)
	}

	var prevPipeReadFd int = -1
	var processes []int

	// syscall.Pipe() creates a unidirectional data channel in the kernel.
	// It returns two file descriptors:
	//   pipeFd[0] → read end (read-only)
	//   pipeFd[1] → write end (write-only)
	//
	// Even though both values are just `int`s in userspace, the kernel
	// enforces read/write permissions by associating them with *different*
	// `struct file` objects internally:
	//
	//   - pipeFd[0] maps to a struct file with FMODE_READ set
	//   - pipeFd[1] maps to a struct file with FMODE_WRITE set
	//
	// Both ultimately reference the same kernel pipe buffer, but attempts
	// to misuse them will fail:
	//   - write(pipeFd[0], ...) → EBADF (because it's read-only)
	//   - read(pipeFd[1], ...)  → EBADF (because it's write-only)
	//
	// Why two FDs instead of one?
	//   - This makes pipes "half-duplex": one side writes, the other reads.
	//   - It prevents processes from accidentally reading their own writes.
	//   - In shell pipelines, it allows passing exactly the write-end to cmd1
	//     and the read-end to cmd2, which matches the model `cmd1 | cmd2`.
	//
	// If you want *bidirectional* communication, use `socketpair()` instead:
	// both FDs can read/write in that case.
	var cmdStderrFd uintptr = os.Stderr.Fd()
	for i, cmd := range p.Commands {
		var pipeFd [2]int
		if i < len(p.Commands)-1 {
			if err := syscall.Pipe(pipeFd[:]); err != nil {
				fmt.Println("pipe failed: %v\n", err)
				return
			}
		}

		cmdStdinFd := os.Stdin.Fd()
		if prevPipeReadFd != -1 {
			cmdStdinFd = uintptr(prevPipeReadFd)
		}

		var cmdStdoutFd uintptr
		if i < len(p.Commands)-1 {
			cmdStdoutFd = uintptr(pipeFd[1])
		} else {
			cmdStdoutFd = stdoutFdPipe
			cmdStderrFd = stderrFdPipe
		}

		// always forks -> subshells
		pid, cmdFdsToClose, err := runCommandWithFds(cmd, cmdStdinFd, cmdStdoutFd, cmdStderrFd)

		if err != nil {
			fmt.Printf("failed to run command: %v\n", err)
			return
		}

		// Close redirect FDs in parent (they are only used by the child)
		for _, fd := range cmdFdsToClose {
			syscall.Close(fd)
		}

		if pid > 0 {
			processes = append(processes, pid)
		}

		// close unused FDs
		if prevPipeReadFd != -1 {
			syscall.Close(prevPipeReadFd)
		}
		if i < len(p.Commands)-1 {
			syscall.Close(pipeFd[1])
			prevPipeReadFd = pipeFd[0]
		}
	}

	// wait for all children
	for _, pid := range processes {
		var status syscall.WaitStatus
		syscall.Wait4(pid, &status, 0, nil)
	}

}

func runCommandWithFds(cmd ast.SimpleCommand, stdinFd, stdoutFd, stderrFd uintptr) (pid int, fdsToClose []int, err error) {
	redirectStdoutNullable, redirectStderrNullable, err := setupRedirectsFd(cmd.Redirects)
	if err != nil {
		return -1, nil, err
	}

	if redirectStdout, hasValue := redirectStdoutNullable.Get(); hasValue {
		stdoutFd = uintptr(redirectStdout)
		fdsToClose = append(fdsToClose, redirectStdout)
	}

	if redirectStderr, hasValue := redirectStderrNullable.Get(); hasValue {
		stderrFd = uintptr(redirectStderr)
		fdsToClose = append(fdsToClose, redirectStderr)
	}

	cmdName := cmd.Args[0]

	// Handle builtins inside pipelines by running them in a subshell.
	//
	// In a pipeline, builtins need to behave like external commands.
	// Normally, builtins (cd, exit, pwd, type) run inside the shell process.
	// But in a pipeline (e.g., `pwd | grep home`):
	//   - The builtin must write to the pipe so the next command can read from it.
	//   - Simply running the builtin in the parent shell would be problematic:
	//       * stdout might still be connected to the terminal or another destination,
	//         so writing directly could break the pipeline flow.
	//       * The parent shell would block until the builtin finishes, breaking
	//         concurrent execution expected in pipelines.
	//       * Any state changes (e.g., cd, exit, export) would affect the parent shell
	//         prematurely, which is incorrect.
	// To solve this, we fork a child process and run the builtin there:
	//   - The child inherits the proper pipe file descriptors for stdin, stdout, and stderr.
	//   - The child executes the builtin logic in isolation.
	//   - The parent shell remains unaffected, and the pipeline executes correctly.
	//
	// This ensures builtins integrate seamlessly into pipelines, just like external commands.
	switch cmdName {
	case "exit", "cd", "pwd", "type":
		// ForkExec explanation:
		//
		// ForkExec is a low-level system call in Go that combines two classic Unix steps:
		//   1. fork()  -> create a new child process, which is a copy of the parent
		//   2. exec()  -> replace the child process memory with a new program
		//
		// In a shell, ForkExec is used to run commands (external or builtins in a pipeline)
		// in a **separate process** (a subprocess) so that:
		//   - The parent shell remains unaffected (its working directory, environment, and FDs stay intact)
		//   - Stdout, stdin, and stderr can be redirected independently for pipes and files
		//   - Builtins like `pwd` can behave like external commands when part of a pipeline
		//
		// Parameters:
		//   - argv0: path to the executable (e.g., "/proc/self/exe" for builtin reruns)
		//   - argv: arguments for the new program
		//   - attr: ProcAttr specifying environment and file descriptor setup
		//
		// Returns:
		//   - pid: process ID of the child
		//   - err: any error that occurred during fork or exec
		//
		// In short, ForkExec allows the shell to run any command or builtin in an isolated
		// subprocess, connect pipes and redirects correctly, and maintain proper shell state.
		pid, err := syscall.ForkExec("/proc/self/exe",
			cmd.Args,
			&syscall.ProcAttr{
				Env:   os.Environ(),
				Files: []uintptr{stdinFd, stdoutFd, stderrFd},
			})
		if err != nil {
			return -1, fdsToClose, fmt.Errorf("fork builtin failed: %w", err)
		}
		return pid, fdsToClose, nil
	}

	// External program
	pid, err = runExecutableWithFds(cmd, stdinFd, stdoutFd, stderrFd)
	return pid, fdsToClose, err
}

func setupRedirectsFd(redirects []ast.Redirect) (stdoutFd, stderrFd nullable.Nullable[int], err error) {
	// Start with nil, meaning default stdout/stderr
	var currentStdout, currentStderr nullable.Nullable[int]

	for _, r := range redirects {
		var err error
		switch r.Type {
		case ast.RedirectStdout:
			if err = openAndReplace(&currentStdout, r.Target, flagReadWriteCreate); err != nil {
				return currentStdout, currentStderr, fmt.Errorf("redirect setup failed: %w", err)
			}
		case ast.RedirectStdoutAppend:
			if err = openAndReplace(&currentStdout, r.Target, flagReadWriteCreateAppend); err != nil {
				return currentStdout, currentStderr, fmt.Errorf("redirect setup failed: %w", err)
			}
		case ast.RedirectStderr:
			if err = openAndReplace(&currentStderr, r.Target, flagReadWriteCreate); err != nil {
				return currentStdout, currentStderr, fmt.Errorf("redirect setup failed: %w", err)
			}
		case ast.RedirectStderrAppend:
			if err = openAndReplace(&currentStderr, r.Target, flagReadWriteCreateAppend); err != nil {
				return currentStdout, currentStderr, fmt.Errorf("redirect setup failed: %w", err)
			}
		}
	}

	return currentStdout, currentStderr, nil
}

func openAndReplace(currentFd *nullable.Nullable[int], path string, flags int) error {
	fd, err := openRedirectFile(path, flags)
	if err != nil {
		return err
	}

	// Close previous FD if it exists
	if val, hasValue := currentFd.Get(); hasValue {
		syscall.Close(val)
	}

	// Set the new FD
	currentFd.Set(fd)

	return nil
}

// File permissions (read/write for owner, read for others)
const defaultFilePerm = 0o644

// Precomputed flags for efficiency
var (
	flagReadWriteCreate       = syscall.O_RDWR | syscall.O_CREAT
	flagReadWriteCreateAppend = syscall.O_RDWR | syscall.O_CREAT | syscall.O_APPEND
)

// openRedirectFile wraps syscall.Open with default permissions
func openRedirectFile(path string, flags int) (int, error) {
	fd, err := syscall.Open(path, flags, defaultFilePerm)
	if err != nil {
		return 0, err
	}
	return fd, nil
}
