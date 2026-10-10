// formatjn is built as the _format_jn command: a personal C/C++ blank-line formatter.
//
// Usage: formatjn [file|dir ...]   format files in place (dirs are searched recursively)
//
//	formatjn                  format the current directory; if stdin is piped,
//	                          read stdin and write stdout instead
//
// See format.go for the formatting rules. Prints a timing summary when done (to stderr when
// filtering stdin, so the output stays clean). Files are formatted in parallel; the report is
// printed in path order.
package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"commandtools/internal/skipdir"
)

func summary(start time.Time, files, formatted int) string {
	ms := float64(time.Since(start).Microseconds()) / 1000
	return fmt.Sprintf("Done: %d file(s), %d formatted, in %.1f ms", files, formatted, ms)
}

// joinPath mirrors Python's os.path.join so reported paths look the same as the Python script's.
func joinPath(dir, name string) string {
	if dir == "" || strings.HasSuffix(dir, "/") || strings.HasSuffix(dir, string(os.PathSeparator)) {
		return dir + name
	}

	return dir + string(os.PathSeparator) + name
}

// walk visits matching files of dir (files first, then subdirectories, both sorted).
func walk(dir string, visit func(string)) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var subdirs []string
	for _, e := range entries {
		switch {
		case e.IsDir():
			if !skipdir.Skip(e.Name()) {
				subdirs = append(subdirs, e.Name())
			}
		case extensions[fileExt(e.Name())]:
			visit(joinPath(dir, e.Name()))
		}
	}

	for _, s := range subdirs {
		walk(joinPath(dir, s), visit)
	}
}

func expand(args []string) []string {
	var paths []string
	for _, arg := range args {
		if info, err := os.Stat(arg); err == nil && info.IsDir() {
			walk(arg, func(p string) { paths = append(paths, p) })
		} else {
			paths = append(paths, arg)
		}
	}

	return paths
}

type result struct {
	skipped, formatted bool
	err                error
}

func process(path string) result {
	if !extensions[fileExt(path)] {
		return result{skipped: true}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return result{err: err}
	}

	text := string(data)
	formatted := formatText(text)
	if formatted == text {
		return result{}
	}

	if err := os.WriteFile(path, []byte(formatted), 0o644); err != nil {
		return result{err: err}
	}

	return result{formatted: true}
}

func stdinIsPiped() bool {
	info, err := os.Stdin.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice == 0
}

func run(args []string, stdin io.Reader, stdinPiped bool, stdout, stderr io.Writer) int {
	start := time.Now()
	if len(args) == 0 {
		if stdinPiped {
			data, err := io.ReadAll(stdin)
			if err != nil {
				fmt.Fprintln(stderr, "error:", err)
				return 1
			}

			io.WriteString(stdout, formatText(string(data)))
			fmt.Fprintln(stderr, summary(start, 1, 1))
			return 0
		}

		args = []string{"."}
	}

	paths := expand(args)
	results := make([]result, len(paths))
	workers := runtime.GOMAXPROCS(0)
	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				results[i] = process(paths[i])
			}
		}()
	}

	for i := range paths {
		jobs <- i
	}

	close(jobs)
	wg.Wait()

	files, formatted, code := 0, 0, 0
	for i, r := range results {
		switch {
		case r.skipped:
			fmt.Fprintf(stdout, "Skipped %s\n", paths[i])
		case r.err != nil:
			files++
			fmt.Fprintf(stderr, "error: %v\n", r.err)
			code = 1
		default:
			files++
			if r.formatted {
				formatted++
				fmt.Fprintf(stdout, "Formatted %s\n", paths[i])
			}
		}
	}

	fmt.Fprintln(stdout, summary(start, files, formatted))
	return code
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, stdinIsPiped(), os.Stdout, os.Stderr))
}
