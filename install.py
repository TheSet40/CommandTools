#!/usr/bin/env python3
"""Put the commandtools scripts on PATH (Windows and Linux/macOS).

Usage: install            add to PATH
       install -u         remove from PATH (also --uninstall)

Creates a generated bin/ folder with one launcher per top-level *.py script
(name.cmd on Windows, a symlink on Linux) and adds only that folder to PATH.
Re-run after adding scripts or moving the repo.
"""
import os
import shutil
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
BIN = ROOT / "bin"
MARK = "# commandtools"
WIN = os.name == "nt"
WIN_REG_LIMIT = 32767   # max length of a registry string value
WIN_LEGACY_LIMIT = 2047  # setx / older tools truncate or choke above this


def scripts():
    return sorted(p for p in ROOT.glob("*.py") if p.name != Path(__file__).name)


def make_launchers():
    BIN.mkdir(exist_ok=True)
    for s in scripts():
        if WIN:
            (BIN / f"{s.stem}.cmd").write_text(f'@"{sys.executable}" "{s}" %*\r\n')
        else:
            link = BIN / s.stem
            if link.is_symlink() or link.exists():
                link.unlink()
            link.symlink_to(s)
            s.chmod(s.stat().st_mode | 0o111)
    print(f"Launchers in {BIN}: " + ", ".join(s.stem for s in scripts()))


def windows_path(add):
    import winreg
    import ctypes

    def norm(p):
        return os.path.normcase(os.path.normpath(os.path.expandvars(p)))

    with winreg.OpenKey(winreg.HKEY_CURRENT_USER, "Environment", 0,
                        winreg.KEY_READ | winreg.KEY_SET_VALUE) as key:
        try:
            current, kind = winreg.QueryValueEx(key, "Path")
        except FileNotFoundError:
            current, kind = "", winreg.REG_EXPAND_SZ
        parts = current.split(";")
        present = any(norm(p) == norm(str(BIN)) for p in parts)

        if add and present:
            print(f"Already in PATH: {BIN}")
            return
        if not add and not present:
            print(f"Not in PATH: {BIN}")
            return
        if add:
            new = current + ("" if not current or current.endswith(";") else ";") + str(BIN)
        else:
            new = ";".join(p for p in parts if not p or norm(p) != norm(str(BIN)))
        if len(new) > WIN_REG_LIMIT:
            sys.exit(f"ERROR: PATH would be {len(new)} chars, over the {WIN_REG_LIMIT} registry limit. Not changed.")
        if len(new) > WIN_LEGACY_LIMIT:
            print(f"WARNING: PATH is {len(new)} chars; some older tools only handle {WIN_LEGACY_LIMIT}.")
        # Keep REG_EXPAND_SZ so %VAR% entries keep working.
        winreg.SetValueEx(key, "Path", 0, kind if kind in (winreg.REG_SZ, winreg.REG_EXPAND_SZ)
                          else winreg.REG_EXPAND_SZ, new)
    ctypes.windll.user32.SendMessageTimeoutW(0xFFFF, 0x1A, 0, "Environment", 2, 5000, ctypes.byref(ctypes.c_ulong()))
    print(("Added to" if add else "Removed from") + f" user PATH: {BIN}")
    if add:
        print("Open a new terminal to pick it up.")


def rc_files():
    home = Path.home()
    existing = [home / n for n in (".bashrc", ".zshrc") if (home / n).exists()]
    return existing or [home / ".profile"]


def linux_path(add):
    line = f'export PATH="$PATH:{BIN}"  {MARK}'
    files = rc_files() if add else [p for p in (Path.home() / n for n in (".bashrc", ".zshrc", ".profile")) if p.exists()]
    for rc in files:
        text = rc.read_text() if rc.exists() else ""
        lines = [l for l in text.splitlines() if not l.rstrip().endswith(MARK)]
        has = len(lines) != len(text.splitlines())
        if add and has:
            print(f"Already in {rc}")
            continue
        if add:
            lines.append(line)
        elif not has:
            continue
        rc.write_text("\n".join(lines) + "\n")
        print(("Added to " if add else "Removed from ") + str(rc))
    if add:
        print("Open a new terminal (or `source` your rc file) to pick it up.")


def main():
    args = sys.argv[1:]
    if any(a in ("-h", "--help", "/?") for a in args):
        print(__doc__)
        return
    uninstall = any(a in ("-u", "--uninstall") for a in args)
    if not uninstall:
        make_launchers()
    (windows_path if WIN else linux_path)(not uninstall)
    if uninstall and BIN.exists():
        shutil.rmtree(BIN)


if __name__ == "__main__":
    main()
