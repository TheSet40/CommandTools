#!/usr/bin/env python3
"""Count real code lines (no blanks, no comment-only lines) in source files.

Usage: codelines [directory] [--summary] [--shortsummary] [--chars]
  --summary       only print the summary, not each file
  --shortsummary  summary as a single total line instead of the per-language table
  --chars         also show characters per file and in the short summary
                  (counted code lines, without indentation)

Only language files are counted; data/docs (xml, json, md, yaml ...) are ignored.
Physical lines are counted, so editor soft-wrapping has no effect.
"""
import os
import sys

# extension -> (language, comment style)
C = "c"        # // and /* */
HASH = "hash"  # #
DASH = "dash"  # --
REM = "rem"    # rem / ::
LANGS = {}
for exts, lang, style in [
    ((".h", ".hpp", ".c", ".cpp", ".cc", ".cxx"), "C/C++", C),
    ((".cs",), "C#", C), ((".java",), "Java", C), ((".kt",), "Kotlin", C), ((".swift",), "Swift", C),
    ((".m", ".mm"), "Objective-C", C), ((".go",), "Go", C), ((".rs",), "Rust", C), ((".dart",), "Dart", C),
    ((".php",), "PHP", "php"), ((".js", ".jsx", ".mjs"), "JavaScript", C), ((".ts", ".tsx"), "TypeScript", C),
    ((".vue",), "Vue", C), ((".glsl", ".hlsl", ".shader"), "Shader", C), ((".css", ".scss"), "CSS", C),
    ((".py",), "Python", HASH), ((".rb",), "Ruby", HASH), ((".sh",), "Shell", HASH), ((".ps1",), "PowerShell", HASH),
    ((".lua",), "Lua", DASH), ((".sql",), "SQL", DASH), ((".bat", ".cmd"), "Batch", REM),
]:
    for e in exts:
        LANGS[e] = (lang, style)

SKIP_DIRS = {"node_modules", ".git", "bin", "obj", "build", "dist", "vendor", ".dart_tool",
             "Pods", "__pycache__", ".venv", "venv"}


def count_file(path, style):
    n, chars, in_block = 0, 0, False
    with open(path, encoding="utf-8", errors="replace") as f:
        for raw in f:
            s = raw.strip()
            if not s:
                continue
            if style == HASH:
                code = not s.startswith("#")
                n += code
                chars += len(s) * code
                continue
            if style == DASH:
                code = not s.startswith("--")
                n += code
                chars += len(s) * code
                continue
            if style == REM:
                code = not (s.lower().startswith("rem ") or s.lower() == "rem" or s.startswith("::"))
                n += code
                chars += len(s) * code
                continue
            has_code, i = False, 0
            while i < len(s):
                if in_block:
                    j = s.find("*/", i)
                    if j < 0:
                        break
                    in_block, i = False, j + 2
                elif s.startswith("//", i) or (style == "php" and s[i] == "#" and not s.startswith("#[", i)):
                    break
                elif s.startswith("/*", i):
                    in_block, i = True, i + 2
                else:
                    has_code = has_code or not s[i].isspace()
                    i += 1
            n += has_code
            chars += len(s) * has_code
    return n, chars


def sv(n):
    return f"{n:,}".replace(",", " ")


def pct(part, whole):
    return f"{part / whole * 100:.1f}".replace(".", ",") + " %" if whole else "-"


def print_ascii_table(headers, rows, right_aligned=(), footer=None):
    rows = list(rows)
    all_rows = rows + ([footer] if footer else [])
    widths = [
        max(len(str(row[i])) for row in [headers] + all_rows)
        for i in range(len(headers))
    ]

    def border():
        return "+" + "+".join("-" * (width + 2) for width in widths) + "+"

    def row(values):
        cells = []
        for i, value in enumerate(values):
            value = str(value)
            cells.append(value.rjust(widths[i]) if i in right_aligned else value.ljust(widths[i]))
        return "| " + " | ".join(cells) + " |"

    print(border())
    print(row(headers))
    print(border())
    for values in rows:
        print(row(values))
    if footer:
        print(border())
        print(row(footer))
    print(border())


def print_table(by_lang, files, total, total_chars):
    rows = []
    for lang, (file_count, line_count, char_count) in sorted(
        by_lang.items(), key=lambda kv: -kv[1][1]
    ):
        rows.append((
            lang,
            sv(file_count), pct(file_count, files),
            sv(line_count), pct(line_count, total),
            sv(char_count), pct(char_count, total_chars),
        ))
    footer = (
        "Total",
        sv(files), pct(files, files),
        sv(total), pct(total, total),
        sv(total_chars), pct(total_chars, total_chars),
    )
    print_ascii_table(
        ("Language", "Files", "Files %", "Lines", "Lines %", "Chars", "Chars %"),
        rows,
        right_aligned=(1, 2, 3, 4, 5, 6),
        footer=footer,
    )


def main():
    args = sys.argv[1:]
    if any(a in ("-h", "--help", "/?") for a in args):
        print(__doc__)
        sys.exit(1)
    summary = "--summary" in args
    short = "--shortsummary" in args
    show_chars = "--chars" in args
    rest = [a for a in args if a not in ("--summary", "--shortsummary", "--chars", "--perlanguage")]
    if len(rest) > 1 or (rest and rest[0].startswith("-")):
        print(f"Unknown argument(s): {' '.join(rest)}\n")
        print(__doc__)
        sys.exit(1)
    root = os.path.abspath(rest[0] if rest else ".")
    if not os.path.isdir(root):
        print(f"ERROR: Directory not found: {root}")
        sys.exit(2)

    by_lang, files, total, total_chars = {}, 0, 0, 0
    file_rows = []
    for dirpath, dirnames, filenames in os.walk(root):
        dirnames[:] = sorted(d for d in dirnames if d not in SKIP_DIRS)
        for name in sorted(filenames):
            lang_style = LANGS.get(os.path.splitext(name)[1].lower())
            if not lang_style:
                continue
            path = os.path.join(dirpath, name)
            try:
                n, c = count_file(path, lang_style[1])
            except OSError:
                continue
            files += 1
            total += n
            total_chars += c
            entry = by_lang.setdefault(lang_style[0], [0, 0, 0])
            entry[0] += 1
            entry[1] += n
            entry[2] += c
            if not summary:
                file_rows.append((
                    sv(n),
                    *([sv(c)] if show_chars else []),
                    os.path.relpath(path, root),
                ))

    if not summary:
        print_ascii_table(
            ("Lines", *(["Chars"] if show_chars else []), "File"),
            file_rows,
            right_aligned=tuple(range(1 + show_chars)),
        )
        print()
    if short:
        print(f"Total: {sv(total)} lines in {sv(files)} files"
              + (f", {sv(total_chars)} characters" if show_chars else ""))
    else:
        print_table(by_lang, files, total, total_chars)


if __name__ == "__main__":
    main()
