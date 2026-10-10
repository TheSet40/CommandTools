package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)


func lines(s ...string) string { return strings.Join(s, "\n") + "\n" }

func check(t *testing.T, source, expected string) {
	t.Helper()
	if got := formatText(source); got != expected {
		t.Errorf("format mismatch\n--- got ---\n%q\n--- want ---\n%q", got, expected)
	}

	// formatting must be stable
	if got := formatText(expected); got != expected {
		t.Errorf("not idempotent\n--- got ---\n%q\n--- want ---\n%q", got, expected)
	}
}

func unchanged(t *testing.T, source string) {
	t.Helper()
	check(t, source, source)
}

func TestBlankLines(t *testing.T) {
	t.Run("blank before continue", func(t *testing.T) {
		check(t,
			lines("        if (a) {", "            pool = createPool(source);", "            continue;", "        }"),
			lines("        if (a) {", "            pool = createPool(source);", "", "            continue;", "        }"))
	})
	t.Run("blank before comment after closing brace", func(t *testing.T) {
		check(t,
			lines(
				"        if (poolNeedsRebuild(source, pool.config)) {",
				"            pool = createPool(source);",
				"            continue;",
				"        }",
				"        // Assigned whole: a string the same length reuses its storage, so a steady frame allocates nothing.",
				"        pool.config = source;"),
			lines(
				"        if (poolNeedsRebuild(source, pool.config)) {",
				"            pool = createPool(source);",
				"",
				"            continue;",
				"        }",
				"",
				"        // Assigned whole: a string the same length reuses its storage, so a steady frame allocates nothing.",
				"        pool.config = source;"))
	})
	t.Run("block comment after closing brace", func(t *testing.T) {
		check(t, lines("if (a) {", "    b();", "}", "/* note */", "c();"),
			lines("if (a) {", "    b();", "}", "", "/* note */", "c();"))
	})
	t.Run("comment ending in semicolon is not a statement", func(t *testing.T) {
		unchanged(t, lines("// do this;", "return x;"))
		unchanged(t, lines("/* do this;", "   and that; */", "return x;"))
	})
	t.Run("comment before control stays attached", func(t *testing.T) {
		unchanged(t, lines("a();", "// why", "b();"))
		check(t, lines("a();", "return b;"), lines("a();", "", "return b;"))
	})
	t.Run("never splits else or catch", func(t *testing.T) {
		unchanged(t, lines("if (a) {", "    b();", "} else {", "    c();", "}"))
		unchanged(t, lines("try {", "    b();", "} catch (...) {", "    c();", "}"))
	})
	t.Run("first line in block untouched", func(t *testing.T) {
		unchanged(t, lines("void f() {", "    return;", "}"))
	})
	t.Run("do while not split", func(t *testing.T) {
		unchanged(t, lines("do {", "    a();", "} while (b);"))
	})
	t.Run("closing braces chain", func(t *testing.T) {
		unchanged(t, lines("void f() {", "    if (a) {", "        b();", "    }", "}"))
	})
	t.Run("preprocessor after brace", func(t *testing.T) {
		unchanged(t, lines("void f() {", "}", "#endif"))
	})
	t.Run("crlf preserved", func(t *testing.T) {
		if got := formatText("a();\r\nreturn b;\r\n"); got != "a();\r\n\r\nreturn b;\r\n" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("no trailing newline preserved", func(t *testing.T) {
		if got := formatText("a();\nreturn b;"); got != "a();\n\nreturn b;" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("empty input", func(t *testing.T) {
		if got := formatText(""); got != "" {
			t.Errorf("got %q", got)
		}
	})
}

func TestHeaders(t *testing.T) {
	t.Run("one liner unchanged", func(t *testing.T) {
		unchanged(t, "long long issuedFrame() const { return issuing; }\n")
	})
	t.Run("inline method unchanged", func(t *testing.T) {
		unchanged(t, lines(
			"    bool drawsAt(const SubMeshAsset &sub, int lod) const {",
			"        return partDrawnAt(sub.role, sub.variant, sub.lodLevel, defaultVariant, lod);",
			"    }"))
	})
	t.Run("switch cases split", func(t *testing.T) {
		check(t,
			lines(
				"constexpr PartBits partBitsFor(SubMeshDraw draw) {",
				"    switch (draw) {",
				"        case SubMeshDraw::Lit: return PartBits::lit;",
				"        case SubMeshDraw::Unlit: return PartBits::unlit;",
				"        case SubMeshDraw::Additive: return PartBits::additive;",
				"        case SubMeshDraw::Fresnel: return PartBits::fresnel;",
				"    }",
				"    return PartBits::none;",
				"}"),
			lines(
				"constexpr PartBits partBitsFor(SubMeshDraw draw) {",
				"    switch (draw) {",
				"        case SubMeshDraw::Lit:",
				"            return PartBits::lit;",
				"        case SubMeshDraw::Unlit:",
				"            return PartBits::unlit;",
				"        case SubMeshDraw::Additive:",
				"            return PartBits::additive;",
				"        case SubMeshDraw::Fresnel:",
				"            return PartBits::fresnel;",
				"    }",
				"",
				"    return PartBits::none;",
				"}"))
	})
	t.Run("consecutive functions separated", func(t *testing.T) {
		check(t,
			lines(
				"constexpr SceneFeatures operator|(SceneFeatures a, SceneFeatures b) {",
				"    return static_cast<SceneFeatures>(static_cast<unsigned>(a) | static_cast<unsigned>(b));",
				"}",
				"constexpr SceneFeatures operator&(SceneFeatures a, SceneFeatures b) {",
				"    return static_cast<SceneFeatures>(static_cast<unsigned>(a) & static_cast<unsigned>(b));",
				"}",
				"constexpr SceneFeatures operator~(SceneFeatures a) {",
				"    return static_cast<SceneFeatures>(~static_cast<unsigned>(a) & static_cast<unsigned>(SceneFeatures::All));",
				"}"),
			lines(
				"constexpr SceneFeatures operator|(SceneFeatures a, SceneFeatures b) {",
				"    return static_cast<SceneFeatures>(static_cast<unsigned>(a) | static_cast<unsigned>(b));",
				"}",
				"",
				"constexpr SceneFeatures operator&(SceneFeatures a, SceneFeatures b) {",
				"    return static_cast<SceneFeatures>(static_cast<unsigned>(a) & static_cast<unsigned>(b));",
				"}",
				"",
				"constexpr SceneFeatures operator~(SceneFeatures a) {",
				"    return static_cast<SceneFeatures>(~static_cast<unsigned>(a) & static_cast<unsigned>(SceneFeatures::All));",
				"}"))
	})
	t.Run("functions inside namespace still separated", func(t *testing.T) {
		check(t,
			lines("namespace a {", "int f() {", "    return 1;", "}", "int g() {", "    return 2;", "}", "}"),
			lines("namespace a {", "int f() {", "    return 1;", "}", "", "int g() {", "    return 2;", "}", "}"))
	})
	t.Run("namespace close is not a block", func(t *testing.T) {
		unchanged(t, lines("namespace a {", "int x;", "}", "namespace b {", "int y;", "}"))
		unchanged(t, lines("namespace a {", "namespace b {", "int x;", "}", "}"))
		unchanged(t, lines(`extern "C" {`, "int x;", "}", "int y;"))
	})
	t.Run("non namespace close still separates", func(t *testing.T) {
		check(t, lines("struct A {", "    int x;", "};", "struct B {", "    int y;", "};"),
			lines("struct A {", "    int x;", "};", "", "struct B {", "    int y;", "};"))
	})
	t.Run("brace in string or comment ignored", func(t *testing.T) {
		check(t,
			lines("void f() {", `    puts("}");`, "}", "int g() {", "    return 1;", "}"),
			lines("void f() {", `    puts("}");`, "}", "", "int g() {", "    return 1;", "}"))
		unchanged(t, lines("namespace a {", "int x; // }", "}", "namespace b {", "}"))
	})
}

func TestCaseSplitting(t *testing.T) {
	t.Run("default split", func(t *testing.T) {
		check(t, lines("switch (x) {", "    default: return 0;", "}"),
			lines("switch (x) {", "    default:", "        return 0;", "}"))
	})
	t.Run("tab indent uses tab", func(t *testing.T) {
		check(t, "\tcase 1: break;\n", "\tcase 1:\n\t\tbreak;\n")
	})
	t.Run("scope resolution in label", func(t *testing.T) {
		check(t, "case A::B::C: x();\n", "case A::B::C:\n    x();\n")
	})
	t.Run("trailing whitespace removed", func(t *testing.T) {
		check(t, "case 1:   return 2;   \n", "case 1:\n    return 2;\n")
	})
	t.Run("not split when body is brace, comment or empty", func(t *testing.T) {
		unchanged(t, "case 1: {\n")
		unchanged(t, "case 1: // note\n")
		unchanged(t, "case 1:\n")
		unchanged(t, "default:\n")
	})
	t.Run("char literal colon left alone", func(t *testing.T) {
		unchanged(t, "case ':': return 1;\n")
	})
	t.Run("identifier starting with case is not a label", func(t *testing.T) {
		unchanged(t, "caseCount: int = 3;\n")
	})
}

func TestSkipDir(t *testing.T) {
	skipped := []string{"build-consumer", "build_debug", "Build-Release", "cmake-build-debug", "x-build",
		".vs", ".idea", ".git", "node_modules", "Debug.dir", "foo.egg-info", "_deps", "builder", "buildings"}

	kept := []string{"src", "include", "rebuild", "rebuilt-src"}

	for _, n := range skipped {
		if !skipDir(n) {
			t.Errorf("%s should be skipped", n)
		}
	}

	for _, n := range kept {
		if skipDir(n) {
			t.Errorf("%s should be kept", n)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runCLI(t *testing.T, stdin string, piped bool, args ...string) (string, string, int) {
	t.Helper()
	var out, errb bytes.Buffer
	code := run(args, strings.NewReader(stdin), piped, &out, &errb)
	return out.String(), errb.String(), code
}

func TestFileHandling(t *testing.T) {
	t.Run("headers are formatted", func(t *testing.T) {
		for _, ext := range []string{".h", ".hpp"} {
			if !extensions[ext] {
				t.Errorf("%s missing from extensions", ext)
			}
		}

		d := t.TempDir()
		path := filepath.Join(d, "a.h")
		writeFile(t, path, "void f() {\n}\nvoid g() {\n}\n")
		out, _, _ := runCLI(t, "", false, d)
		if !strings.Contains(out, "Formatted") {
			t.Errorf("no Formatted line in %q", out)
		}

		got, _ := os.ReadFile(path)
		if string(got) != "void f() {\n}\n\nvoid g() {\n}\n" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("skip dirs and other extensions", func(t *testing.T) {
		d := t.TempDir()
		writeFile(t, filepath.Join(d, "node_modules", "x.js"), "a();\nreturn b;\n")
		writeFile(t, filepath.Join(d, "notes.txt"), "a();\nreturn b;\n")
		if got := expand([]string{d}); len(got) != 0 {
			t.Errorf("expected nothing, got %v", got)
		}
	})
	t.Run("build variant and hidden dirs skipped", func(t *testing.T) {
		skipped := []string{"build-consumer", "build_debug", "Build-Release", "cmake-build-debug", "x-build",
			".vs", ".idea", ".git", "node_modules", "Debug.dir", "foo.egg-info", "_deps", "builder", "buildings"}

		kept := []string{"src", "include", "rebuild", "rebuilt-src"}

		d := t.TempDir()
		for _, n := range append(append([]string{}, skipped...), kept...) {
			writeFile(t, filepath.Join(d, n, "a.cpp"), "int x;\n")
		}

		var found []string
		for _, p := range expand([]string{d}) {
			found = append(found, filepath.Base(filepath.Dir(p)))
		}

		sort.Strings(found)
		sort.Strings(kept)
		if strings.Join(found, ",") != strings.Join(kept, ",") {
			t.Errorf("found %v, want %v", found, kept)
		}
	})
	t.Run("explicit file with other extension is skipped", func(t *testing.T) {
		d := t.TempDir()
		path := filepath.Join(d, "notes.txt")
		writeFile(t, path, "a();\nreturn b;\n")
		out, _, _ := runCLI(t, "", false, path)
		if !strings.Contains(out, "Skipped "+path) {
			t.Errorf("got %q", out)
		}

		got, _ := os.ReadFile(path)
		if string(got) != "a();\nreturn b;\n" {
			t.Errorf("file was modified: %q", got)
		}
	})
	t.Run("missing file reports error", func(t *testing.T) {
		_, errOut, code := runCLI(t, "", false, filepath.Join(t.TempDir(), "nope.cpp"))
		if code == 0 || !strings.Contains(errOut, "error") {
			t.Errorf("code=%d stderr=%q", code, errOut)
		}
	})
	t.Run("timing in output", func(t *testing.T) {
		d := t.TempDir()
		writeFile(t, filepath.Join(d, "a.cpp"), "int x;\n")
		out, _, _ := runCLI(t, "", false, d)
		if !regexp.MustCompile(`Done: 1 file\(s\), 0 formatted, in \d+(\.\d+)? ms`).MatchString(out) {
			t.Errorf("got %q", out)
		}
	})
	t.Run("stdin timing goes to stderr", func(t *testing.T) {
		out, errOut, _ := runCLI(t, "a();\nreturn b;\n", true)
		if out != "a();\n\nreturn b;\n" {
			t.Errorf("stdout %q", out)
		}

		if !regexp.MustCompile(`Done: .* ms`).MatchString(errOut) {
			t.Errorf("stderr %q", errOut)
		}
	})
	t.Run("report order is stable across workers", func(t *testing.T) {
		d := t.TempDir()
		for _, n := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
			writeFile(t, filepath.Join(d, n+".cpp"), "a();\nreturn b;\n")
		}

		out, _, _ := runCLI(t, "", false, d)
		var names []string
		for _, l := range strings.Split(out, "\n") {
			if strings.HasPrefix(l, "Formatted ") {
				names = append(names, filepath.Base(strings.TrimPrefix(l, "Formatted ")))
			}
		}

		if got := strings.Join(names, ","); got != "a.cpp,b.cpp,c.cpp,d.cpp,e.cpp,f.cpp,g.cpp,h.cpp" {
			t.Errorf("got %s", got)
		}
	})
}
