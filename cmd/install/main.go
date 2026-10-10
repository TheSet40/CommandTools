// install builds the commandtools binaries into bin/ and puts only that folder on PATH (Windows and Linux/macOS).
//
// Usage: go run ./cmd/install        build and add to PATH
//
//	go run ./cmd/install -u     remove from PATH and delete bin/ (also --uninstall)
//
// Re-run after changing a tool or moving the repo.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const usage = `Put the commandtools scripts on PATH (Windows and Linux/macOS).

Usage: install            build the tools into bin/ and add it to PATH
       install -u         remove from PATH (also --uninstall)

Builds one binary per folder in cmd/ (except install) into a generated bin/ folder
and adds only that folder to PATH. Re-run after changing a tool or moving the repo.
`

const (
	winRegLimit    = 32767 // max length of a registry string value
	winLegacyLimit = 2047  // setx / older tools truncate or choke above this
)

// commandNames maps cmd/ folders whose command name is not a valid Go folder name.
var commandNames = map[string]string{"formatjn": "_format_jn"}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}

func tools(root string) []string {
	entries, err := os.ReadDir(filepath.Join(root, "cmd"))
	if err != nil {
		fail("Run install from the repository root: " + err.Error())
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() && e.Name() != "install" {
			names = append(names, e.Name())
		}
	}

	sort.Strings(names)
	return names
}

func commandName(folder string) string {
	if name, ok := commandNames[folder]; ok {
		return name
	}

	return folder
}

func buildTools(root, bin string) {
	if err := os.RemoveAll(bin); err != nil {
		fail(err.Error())
	}

	if err := os.MkdirAll(bin, 0o755); err != nil {
		fail(err.Error())
	}

	var commands []string
	for _, folder := range tools(root) {
		name := commandName(folder)
		file := name
		if runtime.GOOS == "windows" {
			file += ".exe"
		}

		cmd := exec.Command("go", "build", "-o", filepath.Join(bin, file), "./cmd/"+folder)
		cmd.Dir = root
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			fail("go build failed for " + folder + ": " + err.Error())
		}

		commands = append(commands, name)
	}

	fmt.Printf("Launchers in %s: %s\n", bin, strings.Join(commands, ", "))
}

func main() {
	for _, a := range os.Args[1:] {
		if a == "-h" || a == "--help" || a == "/?" {
			fmt.Print(usage + "\n")
			return
		}
	}

	uninstall := false
	for _, a := range os.Args[1:] {
		if a == "-u" || a == "--uninstall" {
			uninstall = true
		}
	}

	root, err := os.Getwd()
	if err != nil {
		fail(err.Error())
	}

	bin := filepath.Join(root, "bin")
	if !uninstall {
		buildTools(root, bin)
	}

	updatePath(bin, !uninstall)
	if uninstall {
		os.RemoveAll(bin)
	}
}
