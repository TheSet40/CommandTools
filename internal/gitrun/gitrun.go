// Package gitrun runs git commands the way pushdev and fullmerge print them: a blank line, then the output.
package gitrun

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func exitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}

	fmt.Fprintln(os.Stderr, err)
	return 1
}

// Run prints a blank line, runs git with inherited streams and exits the process with git's code on failure.
func Run(args ...string) {
	fmt.Println()
	cmd := exec.Command("git", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(exitCode(err))
	}
}

// Capture is Run, but returns git's trimmed stdout instead of showing it.
func Capture(args ...string) string {
	fmt.Println()
	var out bytes.Buffer
	cmd := exec.Command("git", args...)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		os.Exit(exitCode(err))
	}

	return strings.TrimSpace(out.String())
}
