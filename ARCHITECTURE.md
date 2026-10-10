# Architecture

CommandTools is a set of small command-line tools. Each tool is a standalone script in the repository root.

## Layout

| Path | Purpose |
| --- | --- |
| `*.py` | One Python 3 script per command (`codelines`, `pushdev`, `fullmerge`, `checkprocess`, `build_eas`, `_format_jn`). |
| `install.py`, `install.bat` | Generate `bin/` launchers and add `bin/` to PATH. |
| `bin/` | Generated launchers, one per script. Not hand-edited by the installer's users. |
| `formatjn/` | Go module: JSON formatter (own `go.mod`). |
| `codelines_go/` | Go module: port of `codelines.py` with identical flags, tables and JSON output. |
| `tests/` | Python tests. |

## codelines

Both implementations walk the directory top-down (sorted), skip `SKIP_DIRS`, count code lines per extension-specific comment style, and accumulate totals per language and per directory (files directly in the directory).
Output flags select the console tables and the shape of the optional JSON file (`total` always, `languages` or `directories` unless `--shortsummary`, `files` unless `--summary`, `chars` only with `--chars`).
Keep the two implementations in step when changing flags, the extension table or the JSON shape.

## Caveats register

| Where | Why it matters | Severity |
| --- | --- | --- |
| `codelines_go/main.go` | Character totals differ from `codelines.py` by a few characters on large trees (Unicode whitespace trimming differs between Go and Python). | 🟡 |
| `codelines_go/codelines.exe` | Built by hand with `go build`; not wired into `bin/` or the installer, so the `codelines` command still runs the Python version. | 🟡 |
