package skipdir
// Package skipdir decides which directories the tools do not descend into.
package skipdir

import (
	"path"
	"strings"
)

// Lowercase names; matching is case-insensitive.
var names = map[string]bool{
	"node_modules": true, "bin": true, "obj": true, "build": true, "dist": true, "out": true,
	"vendor": true, "pods": true, "__pycache__": true, "venv": true, "target": true,
	"_deps": true, "third_party": true, "external": true,
}

// Build variants (build-consumer, cmake-build-release), hidden dirs (.git, .vs, .idea) and
// CMake/virtualenv leftovers.
var patterns = []string{"build*", "*-build", "*_build", "cmake-build-*", ".*", "*.dir", "*.egg-info"}

// Skip reports whether a directory with this base name should be skipped.
func Skip(name string) bool {
	low := strings.ToLower(name)
	if names[low] {
		return true
	}

	for _, p := range patterns {
		if ok, _ := path.Match(p, low); ok {
			return true
		}
	}

	return false
}
