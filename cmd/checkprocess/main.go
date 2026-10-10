// checkprocess shows which processes use a directory (or inspects one PID) and can stop them.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const usage = `Show which processes use a directory (or inspect one PID); optionally kill them.

Usage: checkprocess <directory|pid> [stop|cut|kill]
  stop  Ask every matching process to exit normally (with cleanup) and wait up to 10s
  cut   Close Usages - terminate every matching process
  kill  Same as cut

Windows: uses Sysinternals handle.exe when found, else scans process modules via PowerShell.
Linux:   scans /proc (cwd, exe, open files, mapped libraries). Run as root to see all processes.
`

const psModuleScan = `
$d = $env:CT_DIR.TrimEnd('\').ToLower()
Get-Process -EA SilentlyContinue | ForEach-Object {
  $p = $_; $hit = $false
  try { if ($p.MainModule.FileName.ToLower().StartsWith($d)) { $hit = $true } } catch {}

  if (-not $hit) { try { foreach ($m in $p.Modules) { if ($m.FileName.ToLower().StartsWith($d)) { $hit = $true; break } } } catch {} }

  if ($hit) { "{0}|{1}" -f $p.Id, $p.Name }
}`

const psPidLookup = `$p = Get-Process -Id $env:CT_PID -EA SilentlyContinue; if ($p) { "{0}|{1}" -f $p.Id, $p.Name }`

var isWindows = os.PathSeparator == '\\'

var handleLine = regexp.MustCompile(`(?m)^(\S.*?)\s+pid:\s*(\d+)`)

func output(name string, env []string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...)
	out, _ := cmd.Output()
	return strings.ReplaceAll(string(out), "\r\n", "\n")
}

func powershell(script string, env ...string) []string {
	out := output("powershell", env, "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script)
	var rows []string
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			rows = append(rows, line)
		}
	}

	return rows
}

func addRows(found map[int]string, rows []string) {
	for _, row := range rows {
		id, name, _ := strings.Cut(row, "|")
		if pid, err := strconv.Atoi(id); err == nil {
			found[pid] = name
		}
	}
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func findHandleExe() string {
	for _, name := range []string{"handle.exe", "handle64.exe"} {
		if found, err := exec.LookPath(name); err == nil {
			return found
		}
	}

	home, _ := os.UserHomeDir()
	systemDrive := os.Getenv("SystemDrive")
	if systemDrive == "" {
		systemDrive = "C:"
	}

	for _, base := range []string{os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)"), systemDrive + "\\", filepath.Join(home, "tools")} {
		if base == "" {
			continue
		}

		for _, sub := range []string{"Sysinternals", ""} {
			for _, exe := range []string{"handle.exe", "handle64.exe"} {
				if p := filepath.Join(base, sub, exe); isFile(p) {
					return p
				}
			}
		}
	}

	return ""
}

func findWindows(target string, pid int) map[int]string {
	handle := findHandleExe()
	found := map[int]string{}

	if pid != 0 {
		addRows(found, powershell(psPidLookup, "CT_PID="+strconv.Itoa(pid)))
		if handle != "" && len(found) > 0 {
			fmt.Println(output(handle, nil, "-accepteula", "-nobanner", "-p", strconv.Itoa(pid)))
		}

		return found
	}

	if handle != "" {
		fmt.Print("  [Method: Sysinternals Handle.exe]\n\n")
		out := output(handle, nil, "-accepteula", "-nobanner", target)
		if trimmed := strings.TrimSpace(out); trimmed != "" {
			fmt.Println(trimmed)
		} else {
			fmt.Println("  No open handles found.")
		}

		for _, m := range handleLine.FindAllStringSubmatch(out, -1) {
			id, _ := strconv.Atoi(m[2])
			found[id] = m[1]
		}

		return found
	}

	fmt.Println("  [Method: PowerShell module scan - exe and loaded DLL paths]")
	fmt.Print("  [Tip: install Sysinternals handle.exe for full file-handle detection]\n\n")
	addRows(found, powershell(psModuleScan, "CT_DIR="+target))
	return found
}

func procUses(pid int, target string) bool {
	base := fmt.Sprintf("/proc/%d", pid)
	var links []string
	for _, name := range []string{"cwd", "exe"} {
		link, err := os.Readlink(base + "/" + name)
		if err != nil {
			return false
		}

		links = append(links, link)
	}

	fds, err := os.ReadDir(base + "/fd")
	if err != nil {
		return false
	}

	for _, fd := range fds {
		if link, err := os.Readlink(base + "/fd/" + fd.Name()); err == nil {
			links = append(links, link)
		}
	}

	maps, err := os.ReadFile(base + "/maps")
	if err != nil {
		return false
	}

	for _, line := range strings.Split(string(maps), "\n") {
		if path, ok := mapsPath(line); ok {
			links = append(links, path)
		}
	}

	for _, l := range links {
		if l == target || strings.HasPrefix(l, target+"/") {
			return true
		}
	}

	return false
}

// mapsPath returns the sixth whitespace-separated field of a /proc/<pid>/maps line, keeping its inner spaces.
func mapsPath(line string) (string, bool) {
	rest := line
	for i := 0; i < 5; i++ {
		rest = strings.TrimLeft(rest, " \t")
		end := strings.IndexAny(rest, " \t")
		if end < 0 {
			return "", false
		}

		rest = rest[end:]
	}

	path := strings.TrimSpace(rest)
	return path, path != ""
}

func linuxName(pid int) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return "?"
	}

	return strings.TrimSpace(string(data))
}

func findLinux(target string, pid int) map[int]string {
	found := map[int]string{}

	if pid != 0 {
		if info, err := os.Stat(fmt.Sprintf("/proc/%d", pid)); err != nil || !info.IsDir() {
			return found
		}

		if exe, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid)); err == nil {
			fmt.Printf("  exe : %s\n", exe)
		}

		if data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid)); err == nil {
			fmt.Printf("  cmd : %s\n", strings.TrimSpace(strings.ReplaceAll(string(data), "\x00", " ")))
		}

		found[pid] = linuxName(pid)
		return found
	}

	fmt.Print("  [Method: /proc scan - cwd, exe, open files, mapped libraries]\n\n")
	entries, _ := os.ReadDir("/proc")
	for _, e := range entries {
		p, err := strconv.Atoi(e.Name())
		if err != nil || p == os.Getpid() || p == os.Getppid() {
			continue
		}

		if procUses(p, target) {
			found[p] = linuxName(p)
		}
	}

	return found
}

func signalProcess(pid int, sig syscall.Signal) bool {
	proc, err := os.FindProcess(pid)
	return err == nil && proc.Signal(sig) == nil
}

func taskkill(args ...string) bool {
	return exec.Command("taskkill", args...).Run() == nil
}

// stop on Windows sends WM_CLOSE to windows only, so console/background processes may refuse.
func stop(pid int) bool {
	if isWindows {
		return taskkill("/PID", strconv.Itoa(pid))
	}

	return signalProcess(pid, syscall.SIGTERM)
}

func alive(pid int) bool {
	if isWindows {
		out := output("tasklist", nil, "/FI", fmt.Sprintf("PID eq %d", pid), "/NH")
		for _, field := range strings.Fields(out) {
			if field == strconv.Itoa(pid) {
				return true
			}
		}

		return false
	}

	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return false
	}

	text := string(data)
	fields := strings.Fields(text[strings.LastIndex(text, ")")+1:])
	return len(fields) > 0 && fields[0] != "Z"
}

func kill(pid int) bool {
	if isWindows {
		return taskkill("/F", "/PID", strconv.Itoa(pid))
	}

	return signalProcess(pid, syscall.SIGKILL)
}

func sortedPids(found map[int]string) []int {
	pids := make([]int, 0, len(found))
	for p := range found {
		pids = append(pids, p)
	}

	sort.Ints(pids)
	return pids
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func realPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}

	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved
	}

	return abs
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "/?" {
		fmt.Print(usage + "\n")
		os.Exit(1)
	}

	arg := args[0]
	mode := ""
	if len(args) > 1 {
		mode = strings.ToLower(args[1])
	}

	closeAll := mode == "cut" || mode == "kill"
	graceful := mode == "stop"

	pid, target := 0, ""
	if isDigits(arg) && !isDir(arg) {
		pid, _ = strconv.Atoi(arg)
	} else {
		target = realPath(arg)
		if !isDir(target) {
			fmt.Printf("\n ERROR: Directory not found: %s\n\n", target)
			os.Exit(2)
		}
	}

	subject := target
	if pid != 0 {
		subject = "PID " + strconv.Itoa(pid)
	}

	fmt.Println("\n -------------------------------------------------------")
	fmt.Printf("  checkprocess  |  %s\n", subject)
	fmt.Println(" -------------------------------------------------------")
	modeText := "Report only"
	if closeAll {
		modeText = "Kill matching processes  [cut|kill]"
	} else if graceful {
		modeText = "Graceful stop  [stop]"
	}

	fmt.Printf("  Mode : %s\n\n", modeText)

	var found map[int]string
	if isWindows {
		found = findWindows(target, pid)
	} else {
		found = findLinux(target, pid)
	}

	if len(found) == 0 {
		fmt.Print("  No matching processes found.\n\n")
		return
	}

	pids := sortedPids(found)
	fmt.Println("\n  PID      Name")
	for _, p := range pids {
		fmt.Printf("  %-8d %s\n", p, found[p])
	}

	if closeAll {
		fmt.Printf("\n  Terminating %d process(es)...\n", len(found))
		for _, p := range pids {
			result := "FAILED"
			if kill(p) {
				result = "OK"
			}

			fmt.Printf("  Stopping PID %d (%s)... %s\n", p, found[p], result)
		}
	} else if graceful {
		fmt.Printf("\n  Asking %d process(es) to exit...\n", len(found))
		var asked []int
		for _, p := range pids {
			if stop(p) {
				asked = append(asked, p)
			} else {
				fmt.Printf("  PID %d (%s): request refused\n", p, found[p])
			}
		}

		deadline := time.Now().Add(10 * time.Second)
		for len(asked) > 0 && time.Now().Before(deadline) {
			var still []int
			for _, p := range asked {
				if alive(p) {
					still = append(still, p)
				}
			}

			asked = still
			if len(asked) > 0 {
				time.Sleep(500 * time.Millisecond)
			}
		}

		for _, p := range asked {
			fmt.Printf("  PID %d (%s) is still running - use 'kill' to force\n", p, found[p])
		}

		if len(asked) == 0 {
			fmt.Println("  Requested processes have exited.")
		}
	}

	fmt.Println()
}
