//go:build !windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const marker = "# commandtools"

func rcFiles(home string) []string {
	var existing []string
	for _, n := range []string{".bashrc", ".zshrc"} {
		if _, err := os.Stat(filepath.Join(home, n)); err == nil {
			existing = append(existing, filepath.Join(home, n))
		}
	}

	if len(existing) == 0 {
		return []string{filepath.Join(home, ".profile")}
	}

	return existing
}

func existingRcFiles(home string) []string {
	var files []string
	for _, n := range []string{".bashrc", ".zshrc", ".profile"} {
		if _, err := os.Stat(filepath.Join(home, n)); err == nil {
			files = append(files, filepath.Join(home, n))
		}
	}

	return files
}

func updatePath(bin string, add bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		fail(err.Error())
	}

	line := fmt.Sprintf("export PATH=\"$PATH:%s\"  %s", bin, marker)
	files := existingRcFiles(home)
	if add {
		files = rcFiles(home)
	}

	for _, rc := range files {
		data, _ := os.ReadFile(rc)
		original := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
		if len(data) == 0 {
			original = nil
		}

		var lines []string
		for _, l := range original {
			if !strings.HasSuffix(strings.TrimRight(l, " \t\r"), marker) {
				lines = append(lines, l)
			}
		}

		has := len(lines) != len(original)
		if add && has {
			fmt.Printf("Already in %s\n", rc)
			continue
		}

		if add {
			lines = append(lines, line)
		} else if !has {
			continue
		}

		if err := os.WriteFile(rc, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
			fail(err.Error())
		}

		if add {
			fmt.Printf("Added to %s\n", rc)
		} else {
			fmt.Printf("Removed from %s\n", rc)
		}
	}

	if add {
		fmt.Println("Open a new terminal (or `source` your rc file) to pick it up.")
	}
}
