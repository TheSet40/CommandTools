"""Tests for _format_jn. Run: python -m unittest discover tests -v"""
import contextlib
import io
import os
import re
import sys
import tempfile
import unittest
from unittest import mock

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
import _format_jn as fmt  # noqa: E402


def lines(s):
    """Dedent-free helper: join lines with \\n and a trailing newline."""
    return "\n".join(s) + "\n"


class FormatCase(unittest.TestCase):
    def check(self, source, expected):
        self.assertEqual(fmt.format_text(source), expected)
        # formatting must be stable
        self.assertEqual(fmt.format_text(expected), expected)

    def unchanged(self, source):
        self.check(source, source)


class BlankLines(FormatCase):
    def test_blank_before_continue(self):
        self.check(
            lines(["        if (a) {", "            pool = createPool(source);", "            continue;", "        }"]),
            lines(["        if (a) {", "            pool = createPool(source);", "", "            continue;", "        }"]),
        )

    def test_blank_before_comment_after_closing_brace(self):
        src = lines([
            "        if (poolNeedsRebuild(source, pool.config)) {",
            "            pool = createPool(source);",
            "            continue;",
            "        }",
            "        // Assigned whole: a string the same length reuses its storage, so a steady frame allocates nothing.",
            "        pool.config = source;",
        ])
        expected = lines([
            "        if (poolNeedsRebuild(source, pool.config)) {",
            "            pool = createPool(source);",
            "",
            "            continue;",
            "        }",
            "",
            "        // Assigned whole: a string the same length reuses its storage, so a steady frame allocates nothing.",
            "        pool.config = source;",
        ])
        self.check(src, expected)

    def test_block_comment_after_closing_brace(self):
        self.check(lines(["if (a) {", "    b();", "}", "/* note */", "c();"]),
                   lines(["if (a) {", "    b();", "}", "", "/* note */", "c();"]))

    def test_comment_ending_in_semicolon_is_not_a_statement(self):
        self.unchanged(lines(["// do this;", "return x;"]))
        self.unchanged(lines(["/* do this;", "   and that; */", "return x;"]))

    def test_comment_before_control_stays_attached(self):
        self.unchanged(lines(["a();", "// why", "b();"]))
        self.check(lines(["a();", "return b;"]), lines(["a();", "", "return b;"]))

    def test_never_splits_else_or_catch(self):
        self.unchanged(lines(["if (a) {", "    b();", "} else {", "    c();", "}"]))
        self.unchanged(lines(["try {", "    b();", "} catch (...) {", "    c();", "}"]))

    def test_first_line_in_block_untouched(self):
        self.unchanged(lines(["void f() {", "    return;", "}"]))

    def test_do_while_not_split(self):
        self.unchanged(lines(["do {", "    a();", "} while (b);"]))

    def test_closing_braces_chain(self):
        self.unchanged(lines(["void f() {", "    if (a) {", "        b();", "    }", "}"]))

    def test_preprocessor_after_brace(self):
        self.unchanged(lines(["void f() {", "}", "#endif"]))

    def test_crlf_preserved(self):
        out = fmt.format_text("a();\r\nreturn b;\r\n")
        self.assertEqual(out, "a();\r\n\r\nreturn b;\r\n")

    def test_no_trailing_newline_preserved(self):
        self.assertEqual(fmt.format_text("a();\nreturn b;"), "a();\n\nreturn b;")

    def test_empty_input(self):
        self.assertEqual(fmt.format_text(""), "")


class Headers(FormatCase):
    def test_one_liner_unchanged(self):
        self.unchanged("long long issuedFrame() const { return issuing; }\n")

    def test_inline_method_unchanged(self):
        self.unchanged(lines([
            "    bool drawsAt(const SubMeshAsset &sub, int lod) const {",
            "        return partDrawnAt(sub.role, sub.variant, sub.lodLevel, defaultVariant, lod);",
            "    }",
        ]))

    def test_switch_cases_split(self):
        src = lines([
            "constexpr PartBits partBitsFor(SubMeshDraw draw) {",
            "    switch (draw) {",
            "        case SubMeshDraw::Lit: return PartBits::lit;",
            "        case SubMeshDraw::Unlit: return PartBits::unlit;",
            "        case SubMeshDraw::Additive: return PartBits::additive;",
            "        case SubMeshDraw::Fresnel: return PartBits::fresnel;",
            "    }",
            "    return PartBits::none;",
            "}",
        ])
        expected = lines([
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
            "}",
        ])
        self.check(src, expected)

    def test_consecutive_functions_separated(self):
        src = lines([
            "constexpr SceneFeatures operator|(SceneFeatures a, SceneFeatures b) {",
            "    return static_cast<SceneFeatures>(static_cast<unsigned>(a) | static_cast<unsigned>(b));",
            "}",
            "constexpr SceneFeatures operator&(SceneFeatures a, SceneFeatures b) {",
            "    return static_cast<SceneFeatures>(static_cast<unsigned>(a) & static_cast<unsigned>(b));",
            "}",
            "constexpr SceneFeatures operator~(SceneFeatures a) {",
            "    return static_cast<SceneFeatures>(~static_cast<unsigned>(a) & static_cast<unsigned>(SceneFeatures::All));",
            "}",
        ])
        expected = lines([
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
            "}",
        ])
        self.check(src, expected)

    def test_functions_inside_namespace_still_separated(self):
        src = lines(["namespace a {", "int f() {", "    return 1;", "}", "int g() {", "    return 2;", "}", "}"])
        expected = lines(["namespace a {", "int f() {", "    return 1;", "}", "", "int g() {", "    return 2;", "}", "}"])
        self.check(src, expected)

    def test_namespace_close_is_not_a_block(self):
        self.unchanged(lines(["namespace a {", "int x;", "}", "namespace b {", "int y;", "}"]))
        self.unchanged(lines(["namespace a {", "namespace b {", "int x;", "}", "}"]))
        self.unchanged(lines(['extern "C" {', "int x;", "}", "int y;"]))

    def test_non_namespace_close_still_separates(self):
        self.check(lines(["struct A {", "    int x;", "};", "struct B {", "    int y;", "};"]),
                   lines(["struct A {", "    int x;", "};", "", "struct B {", "    int y;", "};"]))

    def test_namespace_brace_in_string_or_comment_ignored(self):
        self.check(
            lines(["void f() {", '    puts("}");', "}", "int g() {", "    return 1;", "}"]),
            lines(["void f() {", '    puts("}");', "}", "", "int g() {", "    return 1;", "}"]),
        )
        self.unchanged(lines(["namespace a {", "int x; // }", "}", "namespace b {", "}"]))


class CaseSplitting(FormatCase):
    def test_default_split(self):
        self.check(lines(["switch (x) {", "    default: return 0;", "}"]),
                   lines(["switch (x) {", "    default:", "        return 0;", "}"]))

    def test_tab_indent_uses_tab(self):
        self.check("\tcase 1: break;\n", "\tcase 1:\n\t\tbreak;\n")

    def test_scope_resolution_in_label(self):
        self.check("case A::B::C: x();\n", "case A::B::C:\n    x();\n")

    def test_trailing_whitespace_removed(self):
        self.check("case 1:   return 2;   \n", "case 1:\n    return 2;\n")

    def test_not_split_when_body_is_brace_or_comment_or_empty(self):
        self.unchanged("case 1: {\n")
        self.unchanged("case 1: // note\n")
        self.unchanged("case 1:\n")
        self.unchanged("default:\n")

    def test_char_literal_colon_left_alone(self):
        self.unchanged("case ':': return 1;\n")

    def test_identifier_starting_with_case_is_not_a_label(self):
        self.unchanged("caseCount: int = 3;\n")


class FileHandling(unittest.TestCase):
    def run_main(self, *args):
        out, err = io.StringIO(), io.StringIO()
        with mock.patch.object(sys, "argv", ["_format_jn", *args]), \
                contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
            fmt.main()
        return out.getvalue(), err.getvalue()

    def test_headers_are_formatted(self):
        for ext in (".h", ".hpp"):
            self.assertIn(ext, fmt.EXTENSIONS)
        with tempfile.TemporaryDirectory() as d:
            path = os.path.join(d, "a.h")
            with open(path, "w", encoding="utf-8", newline="") as f:
                f.write("void f() {\n}\nvoid g() {\n}\n")
            out, _ = self.run_main(d)
            self.assertIn("Formatted", out)
            with open(path, encoding="utf-8", newline="") as f:
                self.assertEqual(f.read(), "void f() {\n}\n\nvoid g() {\n}\n")

    def test_skip_dirs_and_other_extensions(self):
        with tempfile.TemporaryDirectory() as d:
            os.makedirs(os.path.join(d, "node_modules"))
            for rel in ("node_modules/x.js", "notes.txt"):
                with open(os.path.join(d, rel), "w") as f:
                    f.write("a();\nreturn b;\n")
            self.assertEqual(list(fmt.expand([d])), [])

    def test_build_variant_and_hidden_dirs_skipped(self):
        skipped = ["build-consumer", "build_debug", "Build-Release", "cmake-build-debug", "x-build",
                   ".vs", ".idea", ".git", "node_modules", "Debug.dir", "foo.egg-info", "_deps"]
        kept = ["src", "include", "rebuild", "builder", "buildings", "rebuilt-src"]
        with tempfile.TemporaryDirectory() as d:
            for name in skipped + kept:
                os.makedirs(os.path.join(d, name))
                with open(os.path.join(d, name, "a.cpp"), "w") as f:
                    f.write("int x;\n")
            found = sorted(os.path.basename(os.path.dirname(p)) for p in fmt.expand([d]))
            self.assertEqual(found, sorted(kept))

    def test_timing_in_output(self):
        with tempfile.TemporaryDirectory() as d:
            with open(os.path.join(d, "a.cpp"), "w") as f:
                f.write("int x;\n")
            out, _ = self.run_main(d)
            self.assertRegex(out, r"Done: 1 file\(s\), 0 formatted, in \d+(\.\d+)? ms")

    def test_stdin_timing_goes_to_stderr(self):
        fake = io.StringIO("a();\nreturn b;\n")
        fake.isatty = lambda: False
        out, err = io.StringIO(), io.StringIO()
        with mock.patch.object(sys, "argv", ["_format_jn"]), mock.patch.object(sys, "stdin", fake), \
                contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
            fmt.main()
        self.assertEqual(out.getvalue(), "a();\n\nreturn b;\n")
        self.assertRegex(err.getvalue(), r"Done: .* ms")


if __name__ == "__main__":
    unittest.main()
