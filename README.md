# CommandTools

Small command-line tools for a better workflow. They are Python 3 scripts and run on Windows and Linux.

## Install

```
install.bat        (Windows)
python3 install.py (Linux/macOS)
```

This creates a generated `bin/` folder with one launcher per script and adds only that folder to your PATH.
On Windows it edits the user PATH in the registry, which keeps `%VAR%` entries working and allows up to 32767 characters.
It warns above 2047 characters, which older tools can't handle.
On Linux it appends a marked line to `~/.bashrc` / `~/.zshrc` (or `~/.profile`).
Open a new terminal afterwards. Re-run after adding scripts or moving the repo.

Uninstall with `install.bat -u` (Windows) or `python3 install.py -u` (Linux).

## Tools

| Command | Description |
| --- | --- |
| `checkprocess <dir\|pid> [cut]` | Show which processes use a directory (or inspect one PID). `cut` terminates them. Windows uses Sysinternals `handle.exe` if found, otherwise a module scan. Linux scans `/proc`. Run as admin/root for full results. |
| `codelines [dir] [--summary] [--perlanguage]` | Count real code lines (no blanks or comment-only lines) in source files only (`.h .cpp .cs .php .js .ts .dart .py ...`). Data and docs such as xml, json and md are ignored. |
| `pushdev <branch\|/> <message>` | Pull, add all, create the branch (`/` = stay on current), commit and push. |
| `fullmerge <new branch> <target> <message>` | Pull, add all, create a branch, commit, push it, then merge it into the target. `fullmerge -s` shows status. |
| `build_eas [ios\|manual [st]]` | Run `eas build` for android (default) or ios, or a local gradle `assembleRelease` with `manual`. |
