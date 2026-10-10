# codelines: --perdirectory, --jsonoutput, run time, Go port

Complexity: ⚠️ Intermediate

- [x] Add `--perdirectory` to `codelines.py`
  - Validation: `python codelines.py . --perdirectory` prints a Directory table.
- [x] Add `--jsonoutput <file>` with a structure that follows the flags
  - Validation: JSON has `directories` with `--perdirectory`, no `files` with `--summary`.
- [x] Print the run time (and `runtimeSeconds` in the JSON)
  - Validation: output ends with `Run time: X.XXX s`.
- [x] Write a Go port in `codelines_go/`
  - Validation: totals match the Python script on `e:\dev` (lines and files equal).
- [x] Apply AGENTS.md: camelCase names, README row, ARCHITECTURE.md, caveats register
  - Validation: no remaining snake_case in new code.
