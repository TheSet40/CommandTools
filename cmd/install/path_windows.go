//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"unsafe"
)

var (
	regValueLine = regexp.MustCompile(`(?m)^\s+Path\s+(REG_\w+)\s*(.*?)\r?$`)
	envVar       = regexp.MustCompile(`%([^%]+)%`)
)

func normalize(path string) string {
	expanded := envVar.ReplaceAllStringFunc(path, func(m string) string {
		if value, ok := os.LookupEnv(m[1 : len(m)-1]); ok {
			return value
		}

		return m
	})
	return strings.ToLower(filepath.Clean(expanded))
}

func readUserPath() (string, string) {
	out, err := exec.Command("reg", "query", `HKCU\Environment`, "/v", "Path").Output()
	if err != nil {
		return "", "REG_EXPAND_SZ"
	}

	m := regValueLine.FindStringSubmatch(string(out))
	if m == nil {
		return "", "REG_EXPAND_SZ"
	}

	return m[2], m[1]
}

func broadcastEnvironmentChange() {
	proc := syscall.NewLazyDLL("user32.dll").NewProc("SendMessageTimeoutW")
	env, _ := syscall.UTF16PtrFromString("Environment")
	var result uintptr
	proc.Call(0xFFFF, 0x1A, 0, uintptr(unsafe.Pointer(env)), 2, 5000, uintptr(unsafe.Pointer(&result)))
}

func updatePath(bin string, add bool) {
	current, kind := readUserPath()
	parts := strings.Split(current, ";")
	present := false
	for _, p := range parts {
		if p != "" && normalize(p) == normalize(bin) {
			present = true
		}
	}

	if add && present {
		fmt.Printf("Already in PATH: %s\n", bin)
		return
	}

	if !add && !present {
		fmt.Printf("Not in PATH: %s\n", bin)
		return
	}

	var updated string
	if add {
		separator := ";"
		if current == "" || strings.HasSuffix(current, ";") {
			separator = ""
		}

		updated = current + separator + bin
	} else {
		var kept []string
		for _, p := range parts {
			if p == "" || normalize(p) != normalize(bin) {
				kept = append(kept, p)
			}
		}

		updated = strings.Join(kept, ";")
	}

	if len(updated) > winRegLimit {
		fail(fmt.Sprintf("ERROR: PATH would be %d chars, over the %d registry limit. Not changed.", len(updated), winRegLimit))
	}

	if len(updated) > winLegacyLimit {
		fmt.Printf("WARNING: PATH is %d chars; some older tools only handle %d.\n", len(updated), winLegacyLimit)
	}

	// Keep REG_EXPAND_SZ so %VAR% entries keep working.
	if kind != "REG_SZ" {
		kind = "REG_EXPAND_SZ"
	}

	if err := exec.Command("reg", "add", `HKCU\Environment`, "/v", "Path", "/t", kind, "/d", updated, "/f").Run(); err != nil {
		fail("Could not write the user PATH: " + err.Error())
	}

	broadcastEnvironmentChange()
	if add {
		fmt.Printf("Added to user PATH: %s\n", bin)
		fmt.Println("Open a new terminal to pick it up.")
	} else {
		fmt.Printf("Removed from user PATH: %s\n", bin)
	}
}
