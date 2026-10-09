#!/usr/bin/env python3
"""Personal C/C++ blank-line formatter.

Usage: _format_jn [file|dir ...]   format files in place (dirs are searched recursively)
       _format_jn                  format the current directory; if stdin is piped,
                                   read stdin and write stdout instead

Inserts a blank line before a statement when the previous line ends a statement
(';' or '}') and the line is either a control statement (if/for/while/switch/do/
return/try) or follows a closing '}'. Never splits '} else', braceless bodies,
multi-line statements, or the first line inside a block. Comment lines never end a
statement, but a comment after a closing '}' gets a blank line like any other statement.

Also splits one-line 'case X: stmt;' / 'default: stmt;' onto two lines.

Namespace scopes (namespace / extern "C") are transparent: their braces are not blocks, so
the closing '}' of a namespace never forces a blank line after it.

Prints a timing summary when done (to stderr when filtering stdin, so the output stays clean).
"""
import fnmatch
import os
import re
import sys
import time

# Brace-style languages from codelines, minus CSS/Vue.
EXTENSIONS = {
    ".c", ".cpp", ".cc", ".cxx", ".h", ".hpp", ".hh", ".hxx", ".cs", ".java", ".kt", ".swift",
    ".m", ".mm", ".go", ".rs", ".dart", ".php", ".js", ".jsx", ".mjs", ".ts", ".tsx",
    ".glsl", ".hlsl", ".shader",
}

SKIP_DIRS = {"node_modules", "bin", "obj", "build", "dist", "out", "vendor", "Pods",
             "__pycache__", "venv", "target", "_deps", "third_party", "external"}

# Case-insensitive patterns: build variants (build-consumer, build_debug, cmake-build-release),
# hidden dirs (.git, .vs, .idea, .cache, .dart_tool, .venv), and CMake/virtualenv leftovers.
SKIP_PATTERNS = ("build*", "*-build", "*_build", "cmake-build-*", ".*",
                 "*.dir", "*.egg-info")


def skip_dir(name):
    low = name.lower()
    return name in SKIP_DIRS or low in {d.lower() for d in SKIP_DIRS} or any(
        fnmatch.fnmatch(low, p) for p in SKIP_PATTERNS)

CONTROL = re.compile(r"(if|for|while|switch|do|return|continue|exit|try)\b")
# lines that continue the previous '}' rather than start a new statement
CONTINUATION = re.compile(r"(\}|\)|,|else\b|catch\b|while\b|#|\*)")
# a comment line never ends a statement, even if its text happens to end in ';' or '}'
COMMENT = re.compile(r"(//|/\*|\*)")

NAMESPACE = re.compile(r"\s*(?:inline\s+)?namespace\b[^{;]*\{")
EXTERN_BLOCK = re.compile(r'\s*extern\s+"C(?:\+\+)?"\s*\{')
# 'case X: stmt;' or 'default: stmt;' on one line ('::' is never the label's colon)
CASE_ONE_LINER = re.compile(r"^(\s*)((?:case\b.*?[^:]|default))\s*:(?!:)\s*([^\s/{].*?;.*)$")


def split_cases(lines):
    out = []
    for line in lines:
        m = CASE_ONE_LINER.match(line)
        if m and m.group(2).count("'") % 2 == 0 and m.group(2).count('"') % 2 == 0:
            indent = m.group(1)
            out.append(f"{indent}{m.group(2).rstrip()}:")
            out.append(f"{indent}{chr(9) if chr(9) in indent else '    '}{m.group(3).rstrip()}")
        else:
            out.append(line)
    return out


def strip_code(line, in_comment):
    """Return (line without comments/string contents, still inside a block comment)."""
    out = []
    i = 0
    n = len(line)
    while i < n:
        if in_comment:
            j = line.find("*/", i)
            if j < 0:
                return "".join(out), True
            i = j + 2
            in_comment = False
        elif line.startswith("//", i):
            break
        elif line.startswith("/*", i):
            in_comment = True
            i += 2
        elif line[i] in "\"'":
            quote = line[i]
            i += 1
            while i < n and line[i] != quote:
                i += 2 if line[i] == "\\" else 1
            i += 1
        else:
            out.append(line[i])
            i += 1
    return "".join(out), in_comment


def format_text(text):
    eol = "\r\n" if "\r\n" in text else "\n"
    lines = split_cases(text.splitlines())
    out = []
    scopes = []  # one entry per open brace: "ns" or "block"
    in_comment = False
    prev_comment = False
    prev_ns_close = False
    for line in lines:
        cur = line.strip()
        prev = out[-1].strip() if out else ""
        if cur and prev and prev[-1] in ";}" and not prev_comment and not prev_ns_close:
            after_brace = prev.endswith(("}", "};", "});"))
            if CONTROL.match(cur) and not cur.startswith("while") or (
                after_brace and not CONTINUATION.match(cur) and not cur.endswith(":")
            ):
                out.append("")
        out.append(line)

        prev_comment = in_comment or bool(COMMENT.match(cur))
        code, in_comment = strip_code(line, in_comment)
        is_ns = bool(NAMESPACE.match(code) or EXTERN_BLOCK.match(line))
        last = None
        for ch in code:
            if ch == "{":
                scopes.append("ns" if is_ns else "block")
                last = "open"
            elif ch == "}":
                last = scopes.pop() if scopes else "block"
        prev_ns_close = last == "ns"
    return eol.join(out) + (eol if text.endswith("\n") else "")


def expand(args):
    for arg in args:
        if not os.path.isdir(arg):
            yield arg
            continue
        for dirpath, dirnames, filenames in os.walk(arg):
            dirnames[:] = sorted(d for d in dirnames if not skip_dir(d))
            for name in sorted(filenames):
                if os.path.splitext(name)[1].lower() in EXTENSIONS:
                    yield os.path.join(dirpath, name)


def summary(start, files, formatted):
    ms = (time.perf_counter() - start) * 1000
    return f"Done: {files} file(s), {formatted} formatted, in {ms:.1f} ms"


def main():
    start = time.perf_counter()
    args = sys.argv[1:]
    if not args:
        if not sys.stdin.isatty():
            sys.stdout.write(format_text(sys.stdin.read()))
            print(summary(start, 1, 1), file=sys.stderr)
            return
        args = ["."]
    files = formatted = 0
    for path in expand(args):
        if os.path.splitext(path)[1].lower() not in EXTENSIONS:
            print(f"Skipped {path}")
            continue
        files += 1
        with open(path, encoding="utf-8", newline="") as f:
            text = f.read()
        new = format_text(text)
        if new != text:
            with open(path, "w", encoding="utf-8", newline="") as f:
                f.write(new)
            formatted += 1
            print(f"Formatted {path}")
    print(summary(start, files, formatted))


if __name__ == "__main__":
    main()
