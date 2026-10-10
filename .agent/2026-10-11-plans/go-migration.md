# Migrate all scripts to Go

Complexity: ⚠️ Intermediate

Layout: one root module `commandtools`, one `cmd/<tool>/main.go` per tool, shared git helper in `internal/gitrun`.
Binaries are built by the installer into `bin/` under the original command names (`_format_jn`, `build_eas`, ...).

- [x] Root `go.mod`, move `codelines_go/main.go` -> `cmd/codelines`, `formatjn/*.go` -> `cmd/formatjn`
  - Validation: `go build ./...` and `go test ./...` pass.
- [x] Port `pushdev`, `fullmerge` (shared `internal/gitrun`)
  - Validation: usage text, exit codes and printed lines match Python in a scratch git repo.
- [x] Port `build_eas`
  - Validation: `build_eas manual` outside an Expo project prints the same "not found" message and exits 1.
- [x] Port `checkprocess`
  - Validation: report-only run on a directory and on a PID prints the same layout as Python.
- [x] Port `install` (builds binaries into `bin/`, edits PATH, `-u` uninstalls); `install.bat` runs `go run ./cmd/install`
  - Validation: install creates `bin/*.exe`, uninstall removes `bin/`.
- [x] Delete Python scripts, tests, `__pycache__`, stale `.exe` files; update README, ARCHITECTURE.md (layout, caveats), `.gitignore`
  - Validation: no `.py` files remain; docs reference Go only.
