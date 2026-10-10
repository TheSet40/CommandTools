// fullmerge creates a branch, commits, pushes it, then merges it into a target branch.
package main

import (
	"fmt"
	"os"
	"strings"

	"commandtools/internal/gitrun"
)

const usage = `Create a branch, commit, push it, then merge it into a target branch.

Usage: fullmerge <new branch> <target branch> <commit message...>
       fullmerge -s     show git status
`

func main() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "-s" {
		gitrun.Run("status")
		fmt.Print("\nuse fullmerge with: new branchname, target branch and then commit message\n\n")
		return
	}

	if len(args) < 3 {
		fmt.Print(usage + "\n")
		os.Exit(1)
	}

	newBranch, target, message := args[0], args[1], strings.Join(args[2:], " ")

	fmt.Printf("\"%s -> %s\"\n", newBranch, target)
	fmt.Printf("\"%s\"\n", message)
	gitrun.Run("pull")
	gitrun.Run("add", ".")
	gitrun.Run("checkout", "-b", newBranch)
	gitrun.Run("commit", "-m", message)
	gitrun.Run("push", "--set-upstream", "origin", newBranch)
	gitrun.Run("checkout", target)
	gitrun.Run("merge", newBranch)
}
