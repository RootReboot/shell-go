package executer

import (
	"fmt"
	"os"
	"shelly/app/nullable"
	"shelly/app/parser/ast"
	"syscall"
)

func RunPipeline(p ast.Pipeline) {

	if len(p.Commands) == 1 {
		RunSingleCommand(p.Commands[0])
		// if err != nil {
		// 	fmt.Printf("failed to run single command: %v\n", err)
		// }
		return
	}

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
		pipeCreated := false
		if i < len(p.Commands)-1 {
			if err := syscall.Pipe(pipeFd[:]); err != nil {
				fmt.Println("pipe failed: %v", err)
				return
			}
			pipeCreated = true
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
			cleanupPipeline(err, prevPipeReadFd, &pipeFd, pipeCreated, processes)
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
			prevPipeReadFd = pipeFd[0]
			syscall.Close(pipeFd[1])
		}
	}

	// wait for all children
	for _, pid := range processes {
		var status syscall.WaitStatus
		syscall.Wait4(pid, &status, 0, nil)
	}

}

func RunSingleCommand(cmd ast.SimpleCommand) (err error) {
	if len(cmd.Args) == 0 {
		fmt.Println("Didn't find any data in the command")
	}

	// IMPORTANT: use syscall.Stdin/Stdout/Stderr instead of os.Stdin/os.Stdout/os.Stderr.
	//
	// Why?
	// - When the parent shell forks/execs this process (via ForkExec), it explicitly sets up
	//   file descriptors 0, 1, and 2 to point to pipes, files, or terminals depending on the
	//   pipeline/redirect configuration.
	// - In the re-executed child process, those descriptors are *already correct*:
	//     fd 0 → stdin (maybe a pipe read end)
	//     fd 1 → stdout (maybe a pipe write end)
	//     fd 2 → stderr (maybe redirected to a file)
	// - However, Go’s os.Stdin / os.Stdout / os.Stderr objects are initialized once at program
	//   startup (before the fork/exec). After exec, they may still point to the parent’s original
	//   terminal FDs, not the remapped ones provided by the shell.
	//
	// Using syscall.Stdin/Stdout/Stderr ensures we always respect the current process’s actual
	// file descriptor table, which is exactly what the kernel set up for this child.
	//
	// In short: syscall.* reflects the real fd numbers (0,1,2) after exec, while os.* might be stale.
	stdoutFdPipe := uintptr(syscall.Stdout) // fd 1
	stderrFdPipe := uintptr(syscall.Stderr) // fd 2
	stdinFdPipe := uintptr(syscall.Stdin)   // fd 0

	redirectStdoutNullable, redirectStderrNullable, err := setupRedirectsFd(cmd.Redirects)

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

	cmdName := cmd.Args[0]
	args := cmd.Args[1:]

	switch cmdName {
	case "type":
		handleType(args, stdoutFdPipe)
	case "exit":
		successExit := handleExit(args)
		if successExit {
			os.Exit(0)
		}
	case "pwd":
		handlePWD(stdoutFdPipe)
	case "cd":
		handleCd(args)
	case "history":
		handleHistory(args)
	default:
		pid, err := runExecutableWithFds(cmd, stdinFdPipe, stdoutFdPipe, stderrFdPipe)
		if err != nil {
			fmt.Println(err)
		}

		var status syscall.WaitStatus
		_, err = syscall.Wait4(pid, &status, 0, nil)
		if err != nil {
			return fmt.Errorf("wait4 failed: %w", err)
		}
	}

	return nil
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
	case "exit", "cd", "pwd", "type", "history":
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

		argsForChild := make([]string, len(cmd.Args)+1)
		argsForChild[0] = "--run-builtin"
		copy(argsForChild[1:], cmd.Args)

		pid, err := syscall.ForkExec("/proc/self/exe",
			argsForChild,
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

// cleanupPipeline closes open FDs, kills/waits children, and prints the error.
func cleanupPipeline(err error, prevPipeReadFd int, pipeFd *[2]int, pipeCreated bool, processes []int) {
	if err != nil {
		fmt.Printf("failed to run command: %v\n", err)
	}

	// Close current pipe if we created it
	if pipeCreated && pipeFd != nil {
		syscall.Close(pipeFd[0])
		syscall.Close(pipeFd[1])
	}

	// Close leftover read end from previous iteration
	if prevPipeReadFd != -1 {
		syscall.Close(prevPipeReadFd)
	}

	// Terminate and reap already-started child processes
	for _, pid := range processes {
		// Graceful termination first
		syscall.Kill(pid, syscall.SIGTERM)
	}
	for _, pid := range processes {
		var status syscall.WaitStatus
		syscall.Wait4(pid, &status, 0, nil)
	}
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
