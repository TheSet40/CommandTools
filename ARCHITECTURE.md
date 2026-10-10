# Architecture

CommandTools is a set of small command-line tools written in Go. One module (`commandtools`) at the repository root, one `main` package per tool.

## Layout

| Path | Purpose |
| --- | --- |
| `cmd/<tool>/` | One tool per folder: `codelines`, `pushdev`, `fullmerge`, `checkprocess`, `build_eas`, `formatjn` (installed as `_format_jn`), `install`. |
| `internal/gitrun/` | Shared git runner for `pushdev` and `fullmerge` (prints a blank line, runs git, exits with git's code on failure). |
| `cmd/install/` | Builds every `cmd/*` tool except itself into `bin/` and adds `bin/` to PATH (registry on Windows via `reg.exe`, rc files elsewhere). |
| `install.bat` | Runs `go run ./cmd/install` from the repository root. |
| `bin/` | Generated binaries, one per tool, named after the command. Git-ignored. |

Adding a tool: create `cmd/<name>/main.go` and re-run the installer. A folder name that is not a valid command name needs an entry in `commandNames` in `cmd/install/main.go`.

## codelines

Walks the directory top-down (sorted), skips `skipDirs`, counts code lines per extension-specific comment style, and accumulates totals per language and per directory (files directly in the directory).
Output flags select the console tables and the shape of the optional JSON file (`total` always, `languages` or `directories` unless `--shortsummary`, `files` unless `--summary`, `chars` only with `--chars`).

## Caveats register

| Where | Why it matters | Severity |
| --- | --- | --- |
| `cmd/install/path_windows.go` | PATH is read and written through `reg.exe` instead of the registry API to avoid an external dependency; a very long PATH value is passed on the command line. | 🟡 |
| `cmd/install/main.go` | Installing needs Go on the machine and must run from the repository root (`install.bat` does this). | 🟡 |
