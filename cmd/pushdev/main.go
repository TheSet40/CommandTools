// pushdev pulls, stages everything, optionally creates a branch, commits and pushes.
package main

import (
	"fmt"
	"os"
	"strings"

	"commandtools/internal/gitrun"
)

const usage = `Pull, stage everything, (optionally create a branch,) commit and push.

Usage: pushdev <branch|/> <commit message...>
  /   commit and push on the current branch instead of creating a new one
`

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Print(usage + "\n")
		os.Exit(1)
	}

	branch, message := args[0], strings.Join(args[1:], " ")

	gitrun.Run("pull")
	gitrun.Run("add", ".")
	if branch == "/" {
		branch = gitrun.Capture("rev-parse", "--abbrev-ref", "HEAD")
	} else {
		gitrun.Run("checkout", "-b", branch)
	}

	gitrun.Run("commit", "-m", message)
	gitrun.Run("push", "--set-upstream", "origin", branch)
}
