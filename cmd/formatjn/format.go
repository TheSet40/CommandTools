package main

import (
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Formatting rules of _format_jn.

// Brace-style languages from codelines, minus CSS/Vue.
var extensions = map[string]bool{
	".c": true, ".cpp": true, ".cc": true, ".cxx": true, ".h": true, ".hpp": true, ".hh": true,
	".hxx": true, ".cs": true, ".java": true, ".kt": true, ".swift": true, ".m": true, ".mm": true,
	".go": true, ".rs": true, ".dart": true, ".php": true, ".js": true, ".jsx": true, ".mjs": true,
	".ts": true, ".tsx": true, ".glsl": true, ".hlsl": true, ".shader": true,
}

// Lowercase names; matching is case-insensitive.
var skipDirNames = map[string]bool{
	"node_modules": true, "bin": true, "obj": true, "build": true, "dist": true, "out": true,
	"vendor": true, "pods": true, "__pycache__": true, "venv": true, "target": true,
	"_deps": true, "third_party": true, "external": true,
}

// Build variants (build-consumer, cmake-build-release), hidden dirs (.git, .vs, .idea) and
// CMake/virtualenv leftovers.
var skipDirPatterns = []string{"build*", "*-build", "*_build", "cmake-build-*", ".*", "*.dir", "*.egg-info"}

func skipDir(name string) bool {
	low := strings.ToLower(name)
	if skipDirNames[low] {
		return true
	}

	for _, p := range skipDirPatterns {
		if ok, _ := path.Match(p, low); ok {
			return true
		}
	}

	return false
}

// fileExt mirrors Python's os.path.splitext: leading dots are not an extension (".h" has none).
func fileExt(name string) string {
	trimmed := strings.TrimLeft(filepath.Base(name), ".")
	i := strings.LastIndexByte(trimmed, '.')
	if i < 0 {
		return ""
	}

	return strings.ToLower(trimmed[i:])
}

var (
	controlRe      = regexp.MustCompile(`^(?:if|for|while|switch|do|return|continue|exit|try)\b`)
	continuationRe = regexp.MustCompile(`^(?:\}|\)|,|else\b|catch\b|while\b|#|\*)`)
	commentRe      = regexp.MustCompile(`^(?://|/\*|\*)`)
	namespaceRe    = regexp.MustCompile(`^\s*(?:inline\s+)?namespace\b[^{;]*\{`)
	externBlockRe  = regexp.MustCompile(`^\s*extern\s+"C(?:\+\+)?"\s*\{`)
)

func isWordByte(b byte) bool {
	return b == '_' || b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

// splitLines splits on \n, \r\n and lone \r (no trailing empty element).
func splitLines(text string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '\n':
			lines = append(lines, text[start:i])
			start = i + 1
		case '\r':
			lines = append(lines, text[start:i])
			if i+1 < len(text) && text[i+1] == '\n' {
				i++
			}

			start = i + 1
		}
	}

	if start < len(text) {
		lines = append(lines, text[start:])
	}

	return lines
}

// splitCase recognises 'case X: stmt;' / 'default: stmt;' on one line. Written by hand because
// the Python pattern needs lookahead, which RE2 lacks; it follows that pattern's backtracking:
//
//	^(\s*)((?:case\b.*?[^:]|default))\s*:(?!:)\s*([^\s/{].*?;.*)$
func splitCase(line string) (indent, label, body string, ok bool) {
	rest := strings.TrimLeftFunc(line, unicode.IsSpace)
	indent = line[:len(line)-len(rest)]

	// candidate colon positions in rest, in the order the regex would try them
	var colons []int
	switch {
	case strings.HasPrefix(rest, "case") && (len(rest) == 4 || !isWordByte(rest[4])):
		for i := 5; i < len(rest); i++ {
			if rest[i] == ':' && rest[i-1] != ':' {
				colons = append(colons, i)
			}
		}
	case strings.HasPrefix(rest, "default"):
		i := 7
		for i < len(rest) && unicode.IsSpace(rune(rest[i])) {
			i++
		}

		if i < len(rest) && rest[i] == ':' {
			colons = append(colons, i)
		}
	}

	for _, i := range colons {
		if i+1 < len(rest) && rest[i+1] == ':' {
			continue
		}

		tail := strings.TrimLeftFunc(rest[i+1:], unicode.IsSpace)
		first, w := utf8.DecodeRuneInString(tail)
		if tail == "" || first == '/' || first == '{' {
			continue
		}

		if !strings.Contains(tail[w:], ";") {
			continue
		}

		label = strings.TrimRightFunc(rest[:i], unicode.IsSpace)
		if strings.Count(label, "'")%2 != 0 || strings.Count(label, `"`)%2 != 0 {
			return "", "", "", false
		}

		return indent, label, strings.TrimRightFunc(tail, unicode.IsSpace), true
	}

	return "", "", "", false
}

func splitCases(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		indent, label, body, ok := splitCase(line)
		if !ok {
			out = append(out, line)
			continue
		}

		step := "    "
		if strings.Contains(indent, "\t") {
			step = "\t"
		}

		out = append(out, indent+label+":", indent+step+body)
	}

	return out
}

// stripCode returns the line without comments and string contents, and whether it ends inside
// a block comment.
func stripCode(line string, inComment bool) (string, bool) {
	out := make([]byte, 0, len(line))
	n := len(line)
	for i := 0; i < n; {
		switch {
		case inComment:
			j := strings.Index(line[i:], "*/")
			if j < 0 {
				return string(out), true
			}

			i += j + 2
			inComment = false
		case strings.HasPrefix(line[i:], "//"):
			return string(out), inComment
		case strings.HasPrefix(line[i:], "/*"):
			inComment = true
			i += 2
		case line[i] == '"' || line[i] == '\'':
			quote := line[i]
			i++
			for i < n && line[i] != quote {
				if line[i] == '\\' {
					i += 2
				} else {
					i++
				}
			}

			i++
		default:
			out = append(out, line[i])
			i++
		}
	}

	return string(out), inComment
}

func formatText(text string) string {
	eol := "\n"
	if strings.Contains(text, "\r\n") {
		eol = "\r\n"
	}

	lines := splitCases(splitLines(text))
	out := make([]string, 0, len(lines)+len(lines)/8)
	var scopes []bool // one entry per open brace: true for a namespace
	inComment, prevComment, prevNsClose := false, false, false

	for _, line := range lines {
		cur := strings.TrimSpace(line)
		prev := ""
		if len(out) > 0 {
			prev = strings.TrimSpace(out[len(out)-1])
		}

		if cur != "" && prev != "" && !prevComment && !prevNsClose {
			if last := prev[len(prev)-1]; last == ';' || last == '}' {
				afterBrace := strings.HasSuffix(prev, "}") || strings.HasSuffix(prev, "};") ||
					strings.HasSuffix(prev, "});")
				if (controlRe.MatchString(cur) && !strings.HasPrefix(cur, "while")) ||
					(afterBrace && !continuationRe.MatchString(cur) && !strings.HasSuffix(cur, ":")) {
					out = append(out, "")
				}
			}
		}

		out = append(out, line)

		prevComment = inComment || commentRe.MatchString(cur)
		var code string
		code, inComment = stripCode(line, inComment)
		isNs := namespaceRe.MatchString(code) || externBlockRe.MatchString(line)
		last := "" // what the line's final brace did: "open", "ns" (closed namespace) or "block"
		for i := 0; i < len(code); i++ {
			switch code[i] {
			case '{':
				scopes = append(scopes, isNs)
				last = "open"
			case '}':
				last = "block"
				if n := len(scopes); n > 0 {
					if scopes[n-1] {
						last = "ns"
					}

					scopes = scopes[:n-1]
				}
			}
		}

		prevNsClose = last == "ns"
	}

	result := strings.Join(out, eol)
	if strings.HasSuffix(text, "\n") {
		result += eol
	}

	return result
}
