#!/usr/bin/env python3
"""Personal C/C++ blank-line formatter.

Usage: _format_jn [file|dir ...]   format files in place (dirs are searched recursively)
       _format_jn                  format the current directory; if stdin is piped,
                                   read stdin and write stdout instead

Inserts a blank line before a statement when the previous line ends a statement
(';' or '}') and the line is either a control statement (if/for/while/switch/do/
return/try) or follows a closing '}'. Never splits '} else', braceless bodies,
multi-line statements, or the first line inside a block.
"""
import os
import re
import sys

# Brace-style languages from codelines, minus declaration-only files (.h/.hpp) and CSS/Vue.
EXTENSIONS = {
    ".c", ".cpp", ".cc", ".cxx", ".cs", ".java", ".kt", ".swift", ".m", ".mm",
    ".go", ".rs", ".dart", ".php", ".js", ".jsx", ".mjs", ".ts", ".tsx",
    ".glsl", ".hlsl", ".shader",
}

SKIP_DIRS = {"node_modules", ".git", "bin", "obj", "build", "dist", "vendor", ".dart_tool",
             "Pods", "__pycache__", ".venv", "venv"}

CONTROL = re.compile(r"(if|for|while|switch|do|return|continue|exit|try)\b")
# lines that continue the previous '}' rather than start a new statement
CONTINUATION = re.compile(r"(\}|\)|,|else\b|catch\b|while\b|#|\*)")
# a comment line never ends a statement, even if its text happens to end in ';' or '}'
COMMENT = re.compile(r"(//|/\*|\*)")


def format_text(text):
    eol = "\r\n" if "\r\n" in text else "\n"
    lines = text.splitlines()
    out = []
    for line in lines:
        cur = line.strip()
        prev = out[-1].strip() if out else ""
        if cur and prev and prev[-1] in ";}" and not COMMENT.match(prev):
            after_brace = prev.endswith(("}", "};", "});"))
            if CONTROL.match(cur) and not cur.startswith("while") or (
                after_brace and not CONTINUATION.match(cur) and not cur.endswith(":")
            ):
                out.append("")
        out.append(line)
    return eol.join(out) + (eol if text.endswith("\n") else "")


def expand(args):
    for arg in args:
        if not os.path.isdir(arg):
            yield arg
            continue
        for dirpath, dirnames, filenames in os.walk(arg):
            dirnames[:] = sorted(d for d in dirnames if d not in SKIP_DIRS)
            for name in sorted(filenames):
                if os.path.splitext(name)[1].lower() in EXTENSIONS:
                    yield os.path.join(dirpath, name)


def main():
    args = sys.argv[1:]
    if not args:
        if not sys.stdin.isatty():
            sys.stdout.write(format_text(sys.stdin.read()))
            return
        args = ["."]
    for path in expand(args):
        if os.path.splitext(path)[1].lower() not in EXTENSIONS:
            print(f"Skipped {path}")
            continue
        with open(path, encoding="utf-8", newline="") as f:
            text = f.read()
        new = format_text(text)
        if new != text:
            with open(path, "w", encoding="utf-8", newline="") as f:
                f.write(new)
            print(f"Formatted {path}")


if __name__ == "__main__":
    main()
