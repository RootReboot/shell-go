package syscallHelpers

import (
	"fmt"
	"syscall"
)

func WriteWithSyscall(fd int, data []byte) error {
	total := 0
	for total < len(data) {
		n, err := syscall.Write(fd, data[total:])
		if err != nil {
			/*
			   EINTR = "Interrupted system call"

			   - What it means:
			     The write syscall was blocked (e.g., waiting for the fd to be ready),
			     but before it finished, the process caught a signal (like SIGINT,
			     SIGCHLD, SIGALRM, etc.).
			     The kernel aborts the syscall early and returns EINTR instead of completing it.

			   - Why we must handle it:
			     Without retrying, you'd think the write failed, even though nothing
			     is wrong with the file descriptor itself. Retrying is the standard
			     way to let the syscall finish once the signal is handled.

			   - Why it happens:
			     Any blocking syscall can be interrupted by signals. On modern Linux,
			     some syscalls are automatically restarted, but not all. So a write
			     can still return EINTR in edge cases (signals, timers, debugging tools).

			   EAGAIN / EWOULDBLOCK = "Resource temporarily unavailable"

			   - What it means:
			     The write could not proceed immediately because the file descriptor
			     is set to non-blocking mode and the kernel buffer is full.

			   - Why we usually don’t worry about it for terminals:
			     Standard terminal FDs (stdout, stderr) are **blocking by default**, so
			     write calls almost never return EAGAIN. Only if you explicitly set the
			     fd to non-blocking or are writing to a pseudo-terminal/PTY in non-blocking
			     mode could EAGAIN appear. In Go, os.File.Write automatically handles
			     this case for you, so manual handling is rarely needed.

			   Other errors worth knowing for write:
			     - EBADF:
			       Invalid file descriptor (fd closed or never valid).
			     - EPIPE:
			       Writing to a pipe/socket that has no reader. Go will also raise SIGPIPE
			       unless you ignore it.
			     - ENOSPC:
			       No space left on device (for files).
			     - EFAULT (rare):
			       Invalid memory address in buffer (not typical in Go).
			*/
			if err == syscall.EINTR {
				fmt.Println("syscall.Write interrupted (EINTR), retrying...")
				continue // just retry
			}
			return err // propagate other errors
		}
		total += n
	}
	return nil
}
