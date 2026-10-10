// build_eas builds an Expo/EAS app: eas build for android/ios, or a local gradle assembleRelease.
//
// Usage: build_eas            eas build -p android
//
//	build_eas ios        eas build -p ios
//	build_eas manual [st]  local gradle assembleRelease in ./android (st = --stacktrace)
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func run(dir string, name string, args ...string) {
	exe, err := exec.LookPath(name)
	if err != nil {
		fmt.Printf("%s not found in PATH\n", name)
		os.Exit(127)
	}
	cmd := exec.Command(exe, args...)
	cmd.Dir = dir
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err = cmd.Run()
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		os.Exit(0)
	case errors.As(err, &exitErr):
		os.Exit(exitErr.ExitCode())
	default:
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "manual" {
		scriptName := "gradlew"
		if runtime.GOOS == "windows" {
			scriptName = "gradlew.bat"
		}
		gradlew := filepath.Join("android", scriptName)
		if info, err := os.Stat(gradlew); err != nil || info.IsDir() {
			cwd, _ := os.Getwd()
			fmt.Printf("%s not found in %s\n", gradlew, cwd)
			os.Exit(1)
		}
		abs, _ := filepath.Abs(gradlew)
		gradleArgs := []string{"assembleRelease"}
		if len(args) > 1 && args[1] == "st" {
			gradleArgs = append(gradleArgs, "--stacktrace")
		}
		run("android", abs, gradleArgs...)
	}
	platform := "android"
	if len(args) > 0 && args[0] == "ios" {
		platform = "ios"
	}
	run("", "eas", "build", "-p", platform)
}
