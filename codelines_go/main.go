// codelines counts real code lines (no blanks, no comment-only lines) in source files.
// Go port of codelines.py; see usage text below.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const usage = `Usage: codelines [directory] [--summary] [--shortsummary] [--chars] [--perdirectory]
                 [--jsonoutput <file>]
  --summary       only print the summary, not each file
  --shortsummary  summary as a single total line instead of the per-language table
  --chars         also show characters per file and in the short summary
                  (counted code lines, without indentation)
  --perdirectory  summary table per directory (files directly in it) instead of per language
  --jsonoutput    also write the data to <file> as JSON; the structure follows the flags
                  (no "files" with --summary, only "total" with --shortsummary,
                  "directories" instead of "languages" with --perdirectory,
                  "chars" only with --chars)

Only language files are counted; data/docs (xml, json, md, yaml ...) are ignored.
Physical lines are counted, so editor soft-wrapping has no effect.
`

type style int

const (
	styleC style = iota
	styleHash
	styleDash
	styleRem
	stylePHP
)

type lang struct {
	name  string
	style style
}

var langs = map[string]lang{}

func init() {
	for _, d := range []struct {
		exts  []string
		name  string
		style style
	}{
		{[]string{".h", ".hpp", ".c", ".cpp", ".cc", ".cxx"}, "C/C++", styleC},
		{[]string{".cs"}, "C#", styleC}, {[]string{".java"}, "Java", styleC},
		{[]string{".kt"}, "Kotlin", styleC}, {[]string{".swift"}, "Swift", styleC},
		{[]string{".m", ".mm"}, "Objective-C", styleC}, {[]string{".go"}, "Go", styleC},
		{[]string{".rs"}, "Rust", styleC}, {[]string{".dart"}, "Dart", styleC},
		{[]string{".php"}, "PHP", stylePHP}, {[]string{".js", ".jsx", ".mjs"}, "JavaScript", styleC},
		{[]string{".ts", ".tsx"}, "TypeScript", styleC}, {[]string{".vue"}, "Vue", styleC},
		{[]string{".glsl", ".hlsl", ".shader"}, "Shader", styleC}, {[]string{".css", ".scss"}, "CSS", styleC},
		{[]string{".py"}, "Python", styleHash}, {[]string{".rb"}, "Ruby", styleHash},
		{[]string{".sh"}, "Shell", styleHash}, {[]string{".ps1"}, "PowerShell", styleHash},
		{[]string{".lua"}, "Lua", styleDash}, {[]string{".sql"}, "SQL", styleDash},
		{[]string{".bat", ".cmd"}, "Batch", styleRem},
	} {
		for _, e := range d.exts {
			langs[e] = lang{d.name, d.style}
		}
	}
}

var skipDirs = map[string]bool{
	"node_modules": true, ".git": true, "bin": true, "obj": true, "build": true, "dist": true,
	"vendor": true, ".dart_tool": true, "Pods": true, "__pycache__": true, ".venv": true, "venv": true,
}

func countFile(path string, st style) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	n, chars, inBlock := 0, 0, false
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 64*1024*1024)
	sc.Split(splitLines)
	for sc.Scan() {
		s := strings.TrimSpace(sc.Text())
		if s == "" {
			continue
		}
		code := false
		switch st {
		case styleHash:
			code = !strings.HasPrefix(s, "#")
		case styleDash:
			code = !strings.HasPrefix(s, "--")
		case styleRem:
			l := strings.ToLower(s)
			code = !(strings.HasPrefix(l, "rem ") || l == "rem" || strings.HasPrefix(s, "::"))
		default:
			i := 0
			for i < len(s) {
				if inBlock {
					j := strings.Index(s[i:], "*/")
					if j < 0 {
						break
					}
					inBlock, i = false, i+j+2
				} else if strings.HasPrefix(s[i:], "//") ||
					(st == stylePHP && s[i] == '#' && !strings.HasPrefix(s[i:], "#[")) {
					break
				} else if strings.HasPrefix(s[i:], "/*") {
					inBlock, i = true, i+2
				} else {
					r, size := utf8.DecodeRuneInString(s[i:])
					code = code || !unicode.IsSpace(r)
					i += size
				}
			}
		}
		if code {
			n++
			chars += utf8.RuneCountInString(s)
		}
	}
	return n, chars, sc.Err()
}

// splitLines splits on \n, \r\n and lone \r, matching Python's universal newlines.
func splitLines(data []byte, atEOF bool) (int, []byte, error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for i, b := range data {
		if b == '\n' {
			return i + 1, data[:i], nil
		}
		if b == '\r' {
			if i+1 < len(data) {
				if data[i+1] == '\n' {
					return i + 2, data[:i], nil
				}
				return i + 1, data[:i], nil
			}
			if atEOF {
				return i + 1, data[:i], nil
			}
			return 0, nil, nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func sv(n int) string {
	s := strconv.Itoa(n)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(c)
	}
	return b.String()
}

func pct(part, whole int) string {
	if whole == 0 {
		return "-"
	}
	return strings.Replace(fmt.Sprintf("%.1f", float64(part)/float64(whole)*100), ".", ",", 1) + " %"
}

func printTable(headers []string, rows [][]string, right map[int]bool, footer []string) {
	all := append([][]string{headers}, rows...)
	if footer != nil {
		all = append(all, footer)
	}
	widths := make([]int, len(headers))
	for _, r := range all {
		for i, v := range r {
			if w := utf8.RuneCountInString(v); w > widths[i] {
				widths[i] = w
			}
		}
	}
	border := func() string {
		parts := make([]string, len(widths))
		for i, w := range widths {
			parts[i] = strings.Repeat("-", w+2)
		}
		return "+" + strings.Join(parts, "+") + "+"
	}
	row := func(vals []string) string {
		cells := make([]string, len(vals))
		for i, v := range vals {
			pad := strings.Repeat(" ", widths[i]-utf8.RuneCountInString(v))
			if right[i] {
				cells[i] = pad + v
			} else {
				cells[i] = v + pad
			}
		}
		return "| " + strings.Join(cells, " | ") + " |"
	}
	fmt.Println(border())
	fmt.Println(row(headers))
	fmt.Println(border())
	for _, r := range rows {
		fmt.Println(row(r))
	}
	if footer != nil {
		fmt.Println(border())
		fmt.Println(row(footer))
	}
	fmt.Println(border())
}

// group keeps insertion order so ties sort like the Python version (stable).
type group struct {
	keys  []string
	stats map[string]*[3]int
}

func newGroup() *group { return &group{stats: map[string]*[3]int{}} }

func (g *group) add(key string, n, c int) {
	e, ok := g.stats[key]
	if !ok {
		e = &[3]int{}
		g.stats[key] = e
		g.keys = append(g.keys, key)
	}
	e[0]++
	e[1] += n
	e[2] += c
}

func (g *group) sorted() []string {
	keys := append([]string(nil), g.keys...)
	sort.SliceStable(keys, func(i, j int) bool { return g.stats[keys[i]][1] > g.stats[keys[j]][1] })
	return keys
}

func (g *group) printTable(label string, files, total, totalChars int) {
	var rows [][]string
	for _, k := range g.sorted() {
		e := g.stats[k]
		rows = append(rows, []string{k, sv(e[0]), pct(e[0], files), sv(e[1]), pct(e[1], total), sv(e[2]), pct(e[2], totalChars)})
	}
	footer := []string{"Total", sv(files), pct(files, files), sv(total), pct(total, total), sv(totalChars), pct(totalChars, totalChars)}
	printTable([]string{label, "Files", "Files %", "Lines", "Lines %", "Chars", "Chars %"}, rows,
		map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true}, footer)
}

// orderedMap marshals as a JSON object preserving key order.
type orderedMap struct {
	keys []string
	vals map[string]any
}

func (o orderedMap) MarshalJSON() ([]byte, error) {
	var b strings.Builder
	b.WriteByte('{')
	for i, k := range o.keys {
		if i > 0 {
			b.WriteByte(',')
		}
		kb, _ := json.Marshal(k)
		vb, err := json.Marshal(o.vals[k])
		if err != nil {
			return nil, err
		}
		b.Write(kb)
		b.WriteByte(':')
		b.Write(vb)
	}
	b.WriteByte('}')
	return []byte(b.String()), nil
}

func fail(code int, msg string) {
	fmt.Print(msg)
	os.Exit(code)
}

type fileEntry struct {
	path, lang string
	n, c       int
}

func main() {
	start := time.Now()
	args := os.Args[1:]
	var summary, short, showChars, perDir bool
	var jsonPath string
	var rest []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help" || a == "/?":
			fail(1, usage)
		case a == "--summary":
			summary = true
		case a == "--shortsummary":
			short = true
		case a == "--chars":
			showChars = true
		case a == "--perdirectory":
			perDir = true
		case a == "--perlanguage":
		case a == "--jsonoutput":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				fail(1, "--jsonoutput requires a file path\n\n"+usage)
			}
			i++
			jsonPath = args[i]
		case strings.HasPrefix(a, "--jsonoutput="):
			jsonPath = strings.SplitN(a, "=", 2)[1]
		default:
			rest = append(rest, a)
		}
	}
	if len(rest) > 1 || (len(rest) == 1 && strings.HasPrefix(rest[0], "-")) {
		fail(1, fmt.Sprintf("Unknown argument(s): %s\n\n%s", strings.Join(rest, " "), usage))
	}
	rootArg := "."
	if len(rest) == 1 {
		rootArg = rest[0]
	}
	root, _ := filepath.Abs(rootArg)
	if st, err := os.Stat(root); err != nil || !st.IsDir() {
		fail(2, "ERROR: Directory not found: "+root+"\n")
	}

	byLang, byDir := newGroup(), newGroup()
	var files, total, totalChars int
	var fileData []fileEntry

	var walk func(dir string)
	walk = func(dir string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		var subdirs []string
		for _, e := range entries {
			if e.IsDir() {
				if !skipDirs[e.Name()] {
					subdirs = append(subdirs, filepath.Join(dir, e.Name()))
				}
				continue
			}
			l, ok := langs[strings.ToLower(filepath.Ext(e.Name()))]
			if !ok {
				continue
			}
			path := filepath.Join(dir, e.Name())
			n, c, err := countFile(path, l.style)
			if err != nil {
				continue
			}
			files++
			total += n
			totalChars += c
			byLang.add(l.name, n, c)
			relDir, _ := filepath.Rel(root, dir)
			byDir.add(relDir, n, c)
			if !summary {
				rel, _ := filepath.Rel(root, path)
				fileData = append(fileData, fileEntry{rel, l.name, n, c})
			}
		}
		for _, d := range subdirs {
			walk(d)
		}
	}
	walk(root)

	if !summary {
		headers := []string{"Lines"}
		right := map[int]bool{0: true}
		if showChars {
			headers = append(headers, "Chars")
			right[1] = true
		}
		headers = append(headers, "File")
		var rows [][]string
		for _, f := range fileData {
			r := []string{sv(f.n)}
			if showChars {
				r = append(r, sv(f.c))
			}
			rows = append(rows, append(r, f.path))
		}
		printTable(headers, rows, right, nil)
		fmt.Println()
	}
	if short {
		extra := ""
		if showChars {
			extra = ", " + sv(totalChars) + " characters"
		}
		fmt.Printf("Total: %s lines in %s files%s\n", sv(total), sv(files), extra)
	} else if perDir {
		byDir.printTable("Directory", files, total, totalChars)
	} else {
		byLang.printTable("Language", files, total, totalChars)
	}

	if jsonPath != "" {
		stats := func(e [3]int) map[string]any {
			m := map[string]any{"files": e[0], "lines": e[1]}
			if showChars {
				m["chars"] = e[2]
			}
			return m
		}
		statsOrdered := func(e [3]int) orderedMap {
			o := orderedMap{keys: []string{"files", "lines"}, vals: stats(e)}
			if showChars {
				o.keys = append(o.keys, "chars")
			}
			return o
		}
		data := orderedMap{vals: map[string]any{}}
		set := func(k string, v any) {
			data.keys = append(data.keys, k)
			data.vals[k] = v
		}
		set("root", root)
		set("total", statsOrdered([3]int{files, total, totalChars}))
		set("runtimeSeconds", float64(time.Since(start).Microseconds())/1e6)
		if !short {
			g, key := byLang, "languages"
			if perDir {
				g, key = byDir, "directories"
			}
			om := orderedMap{vals: map[string]any{}}
			for _, k := range g.sorted() {
				om.keys = append(om.keys, k)
				om.vals[k] = statsOrdered(*g.stats[k])
			}
			set(key, om)
		}
		if !summary {
			var list []any
			for _, f := range fileData {
				o := orderedMap{keys: []string{"path", "language", "lines"},
					vals: map[string]any{"path": f.path, "language": f.lang, "lines": f.n}}
				if showChars {
					o.keys = append(o.keys, "chars")
					o.vals["chars"] = f.c
				}
				list = append(list, o)
			}
			if list == nil {
				list = []any{}
			}
			set("files", list)
		}
		out, err := json.MarshalIndent(data, "", "  ")
		if err == nil {
			err = os.WriteFile(jsonPath, out, 0o644)
		}
		if err != nil {
			fail(2, "ERROR: "+err.Error()+"\n")
		}
		abs, _ := filepath.Abs(jsonPath)
		fmt.Println("JSON written to " + abs)
	}
	fmt.Printf("Run time: %.3f s\n", time.Since(start).Seconds())
}
