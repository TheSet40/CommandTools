#!/usr/bin/env python3
"""Show which processes use a directory (or inspect one PID); optionally kill them.

Usage: checkprocess <directory|pid> [stop|cut|kill]
  stop  Ask every matching process to exit normally (with cleanup) and wait up to 10s
  cut   Close Usages - terminate every matching process
  kill  Same as cut

Windows: uses Sysinternals handle.exe when found, else scans process modules via PowerShell.
Linux:   scans /proc (cwd, exe, open files, mapped libraries). Run as root to see all processes.
"""
import os
import re
import shutil
import signal
import subprocess
import sys
import time

WIN = os.name == "nt"


def usage():
    print(__doc__)
    sys.exit(1)


def powershell(script, **env):
    out = subprocess.run(
        ["powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script],
        capture_output=True, text=True, env={**os.environ, **env},
    )
    return [l.strip() for l in out.stdout.splitlines() if l.strip()]


def find_handle_exe():
    found = shutil.which("handle.exe") or shutil.which("handle64.exe")
    if found:
        return found
    for base in (os.environ.get("ProgramFiles"), os.environ.get("ProgramFiles(x86)"),
                 os.environ.get("SystemDrive", "C:") + "\\", os.path.join(os.path.expanduser("~"), "tools")):
        if not base:
            continue
        for sub in ("Sysinternals", ""):
            for exe in ("handle.exe", "handle64.exe"):
                p = os.path.join(base, sub, exe)
                if os.path.isfile(p):
                    return p
    return None


def find_windows(target, pid):
    handle = find_handle_exe()
    found = {}
    if pid:
        rows = powershell(
            '$p = Get-Process -Id $env:CT_PID -EA SilentlyContinue; if ($p) { "{0}|{1}" -f $p.Id, $p.Name }',
            CT_PID=str(pid))
        for r in rows:
            i, name = r.split("|", 1)
            found[int(i)] = name
        if handle and found:
            out = subprocess.run([handle, "-accepteula", "-nobanner", "-p", str(pid)],
                                 capture_output=True, text=True).stdout
            print(out)
        return found

    if handle:
        print("  [Method: Sysinternals Handle.exe]\n")
        out = subprocess.run([handle, "-accepteula", "-nobanner", target],
                             capture_output=True, text=True).stdout
        print(out.strip() or "  No open handles found.")
        for m in re.finditer(r"^(\S.*?)\s+pid:\s*(\d+)", out, re.M):
            found[int(m.group(2))] = m.group(1)
        return found

    print("  [Method: PowerShell module scan - exe and loaded DLL paths]")
    print("  [Tip: install Sysinternals handle.exe for full file-handle detection]\n")
    script = r'''
$d = $env:CT_DIR.TrimEnd('\').ToLower()
Get-Process -EA SilentlyContinue | ForEach-Object {
  $p = $_; $hit = $false
  try { if ($p.MainModule.FileName.ToLower().StartsWith($d)) { $hit = $true } } catch {}
  if (-not $hit) { try { foreach ($m in $p.Modules) { if ($m.FileName.ToLower().StartsWith($d)) { $hit = $true; break } } } catch {} }
  if ($hit) { "{0}|{1}" -f $p.Id, $p.Name }
}'''
    for r in powershell(script, CT_DIR=target):
        i, name = r.split("|", 1)
        found[int(i)] = name
    return found


def proc_uses(pid, target):
    base = f"/proc/{pid}"
    try:
        links = [os.readlink(f"{base}/cwd"), os.readlink(f"{base}/exe")]
        for fd in os.listdir(f"{base}/fd"):
            try:
                links.append(os.readlink(f"{base}/fd/{fd}"))
            except OSError:
                pass
        with open(f"{base}/maps", errors="replace") as f:
            links += [l.split(None, 5)[5].strip() for l in f if len(l.split(None, 5)) == 6]
    except OSError:
        return False
    return any(l == target or l.startswith(target + "/") for l in links)


def linux_name(pid):
    try:
        with open(f"/proc/{pid}/comm") as f:
            return f.read().strip()
    except OSError:
        return "?"


def find_linux(target, pid):
    if pid:
        if not os.path.isdir(f"/proc/{pid}"):
            return {}
        try:
            print(f"  exe : {os.readlink(f'/proc/{pid}/exe')}")
        except OSError:
            pass
        try:
            with open(f"/proc/{pid}/cmdline") as f:
                print(f"  cmd : {f.read().replace(chr(0), ' ').strip()}")
        except OSError:
            pass
        return {pid: linux_name(pid)}
    print("  [Method: /proc scan - cwd, exe, open files, mapped libraries]\n")
    me = {os.getpid(), os.getppid()}
    return {int(p): linux_name(p) for p in os.listdir("/proc")
            if p.isdigit() and int(p) not in me and proc_uses(p, target)}


def stop(pid):
    # Windows: WM_CLOSE to windows only, so console/background processes may refuse.
    if WIN:
        return subprocess.run(["taskkill", "/PID", str(pid)], capture_output=True).returncode == 0
    try:
        os.kill(pid, signal.SIGTERM)
        return True
    except OSError:
        return False


def alive(pid):
    if WIN:
        out = subprocess.run(["tasklist", "/FI", f"PID eq {pid}", "/NH"], capture_output=True, text=True).stdout
        return str(pid) in out.split()
    try:
        with open(f"/proc/{pid}/stat") as f:
            return f.read().rsplit(")", 1)[1].split()[0] != "Z"
    except OSError:
        return False


def kill(pid):
    if WIN:
        return subprocess.run(["taskkill", "/F", "/PID", str(pid)], capture_output=True).returncode == 0
    try:
        os.kill(pid, signal.SIGKILL)
        return True
    except OSError:
        return False


def main():
    args = sys.argv[1:]
    if not args or args[0] in ("-h", "--help", "/?"):
        usage()
    arg = args[0]
    mode = args[1].lower() if len(args) > 1 else ""
    close = mode in ("cut", "kill")
    graceful = mode == "stop"

    pid = int(arg) if arg.isdigit() and not os.path.isdir(arg) else 0
    target = "" if pid else os.path.realpath(arg)
    if not pid and not os.path.isdir(target):
        print(f"\n ERROR: Directory not found: {target}\n")
        sys.exit(2)

    print("\n -------------------------------------------------------")
    print(f"  checkprocess  |  {'PID ' + str(pid) if pid else target}")
    print(" -------------------------------------------------------")
    print("  Mode : " + ("Kill matching processes  [cut|kill]" if close
                       else "Graceful stop  [stop]" if graceful else "Report only") + "\n")

    found = find_windows(target, pid) if WIN else find_linux(target, pid)
    if not found:
        print("  No matching processes found.\n")
        return
    print("\n  PID      Name")
    for p, name in sorted(found.items()):
        print(f"  {p:<8} {name}")

    if close:
        print(f"\n  Terminating {len(found)} process(es)...")
        for p, name in sorted(found.items()):
            print(f"  Stopping PID {p} ({name})... " + ("OK" if kill(p) else "FAILED"))
    elif graceful:
        print(f"\n  Asking {len(found)} process(es) to exit...")
        asked = [p for p, name in sorted(found.items())
                 if stop(p) or print(f"  PID {p} ({name}): request refused")]
        deadline = time.time() + 10
        while asked and time.time() < deadline:
            asked = [p for p in asked if alive(p)]
            if asked:
                time.sleep(0.5)
        for p in asked:
            print(f"  PID {p} ({found[p]}) is still running - use 'kill' to force")
        if not asked:
            print("  Requested processes have exited.")
    print()


if __name__ == "__main__":
    main()
