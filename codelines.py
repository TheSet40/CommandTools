#!/usr/bin/env python3
"""Count real code lines (no blanks, no comment-only lines) in source files.

Usage: codelines [directory] [--summary] [--perlanguage]
  --summary      only print the total, not each file
  --perlanguage  print a per-language breakdown

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
    n, in_block = 0, False
    with open(path, encoding="utf-8", errors="replace") as f:
        for raw in f:
            s = raw.strip()
            if not s:
                continue
            if style == HASH:
                n += not s.startswith("#")
                continue
            if style == DASH:
                n += not s.startswith("--")
                continue
            if style == REM:
                n += not (s.lower().startswith("rem ") or s.lower() == "rem" or s.startswith("::"))
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
    return n


def main():
    args = sys.argv[1:]
    if any(a in ("-h", "--help", "/?") for a in args):
        print(__doc__)
        sys.exit(1)
    summary = "--summary" in args
    perlang = "--perlanguage" in args
    rest = [a for a in args if a not in ("--summary", "--perlanguage")]
    if len(rest) > 1 or (rest and rest[0].startswith("-")):
        print(f"Unknown argument(s): {' '.join(rest)}\n")
        print(__doc__)
        sys.exit(1)
    root = os.path.abspath(rest[0] if rest else ".")
    if not os.path.isdir(root):
        print(f"ERROR: Directory not found: {root}")
        sys.exit(2)

    by_lang, files, total = {}, 0, 0
    for dirpath, dirnames, filenames in os.walk(root):
        dirnames[:] = sorted(d for d in dirnames if d not in SKIP_DIRS)
        for name in sorted(filenames):
            lang_style = LANGS.get(os.path.splitext(name)[1].lower())
            if not lang_style:
                continue
            path = os.path.join(dirpath, name)
            try:
                n = count_file(path, lang_style[1])
            except OSError:
                continue
            files += 1
            total += n
            entry = by_lang.setdefault(lang_style[0], [0, 0])
            entry[0] += 1
            entry[1] += n
            if not summary:
                print(f"{n:8}  {os.path.relpath(path, root)}")

    if not summary:
        print()
    if perlang:
        print(f"{'Language':<14}{'Files':>7}{'Lines':>10}")
        for lang, (f, n) in sorted(by_lang.items(), key=lambda kv: -kv[1][1]):
            print(f"{lang:<14}{f:>7}{n:>10}")
        print()
    print(f"Total: {total} lines in {files} files")


if __name__ == "__main__":
    main()
