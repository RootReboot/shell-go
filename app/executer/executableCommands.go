package executer

import (
	"fmt"
	"os"
	"path/filepath"
	"shelly/app/parser/ast"
	"strings"
	"syscall"
)

func runExecutable(cmd ast.SimpleCommand) error {
	if len(cmd.Args) == 0 {
		fmt.Println("empty command")
	}

	cmdName := cmd.Args[0]
	args := cmd.Args[1:]

	binaryPath, err := findExecutableBinaryInPath(cmdName)
	if err != nil {
		return err
	}

	env := os.Environ()

	pid, err := syscall.ForkExec(binaryPath, append([]string{cmdName}, args...), &syscall.ProcAttr{
		Env: env,
		Files: []uintptr{
			os.Stdin.Fd(),
			os.Stdout.Fd(),
			os.Stderr.Fd(),
		},
	})
	if err != nil {
		return fmt.Errorf("fork-exec failed: %w", err)
	}

	var status syscall.WaitStatus
	_, err = syscall.Wait4(pid, &status, 0, nil)
	if err != nil {
		return fmt.Errorf("wait4 failed: %w", err)
	}

	return nil
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

	return "", fmt.Errorf(" %s: command not found", cmd)
}
