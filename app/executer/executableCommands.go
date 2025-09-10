package executer

import (
	"fmt"
	"os"
	"path/filepath"
	"shelly/app/parser/ast"
	"strings"
	"syscall"
)

func runExecutableWithFds(cmd ast.SimpleCommand, stdinFd, stdoutFd, stderrFd uintptr) (pid int, err error) {
	if len(cmd.Args) == 0 {
		fmt.Println("empty command")
	}

	cmdName := cmd.Args[0]
	args := cmd.Args[1:]

	binaryPath, err := findExecutableBinaryInPath(cmdName)
	if err != nil {
		return -1, err
	}

	env := os.Environ()

	pid, err = syscall.ForkExec(binaryPath, append([]string{cmdName}, args...), &syscall.ProcAttr{
		Env: env,
		Files: []uintptr{
			stdinFd,
			stdoutFd,
			stderrFd,
		},
	})

	if err != nil {
		return -1, fmt.Errorf("fork-exec failed: %w", err)
	}

	return
}

func findExecutableBinaryInPath(cmd string) (string, error) {
	pathEnvVar, envVarExists := os.LookupEnv("PATH")
	if !envVarExists {
		return "", fmt.Errorf("PATH environment variable not set")
	}

	pathsToCheck := strings.Split(pathEnvVar, ":")
	for _, path := range pathsToCheck {
		fullPath := filepath.Join(path, cmd)
		fileInfo, err := os.Stat(fullPath)
		if err == nil && fileInfo.Mode().Perm()&0111 != 0 {
			return fullPath, nil
		}
	}

	return "", fmt.Errorf("%s: command not found", cmd)
}
