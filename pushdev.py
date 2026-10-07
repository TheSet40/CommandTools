#!/usr/bin/env python3
"""Pull, stage everything, (optionally create a branch,) commit and push.

Usage: pushdev <branch|/> <commit message...>
  /   commit and push on the current branch instead of creating a new one
"""
import subprocess
import sys


def git(*args, capture=False):
    print()
    r = subprocess.run(["git", *args], capture_output=capture, text=True)
    if r.returncode != 0:
        sys.exit(r.returncode)
    return r.stdout.strip() if capture else None


def main():
    if len(sys.argv) < 3:
        print(__doc__)
        sys.exit(1)
    branch, message = sys.argv[1], " ".join(sys.argv[2:])

    git("pull")
    git("add", ".")
    if branch == "/":
        branch = git("rev-parse", "--abbrev-ref", "HEAD", capture=True)
    else:
        git("checkout", "-b", branch)
    git("commit", "-m", message)
    git("push", "--set-upstream", "origin", branch)


if __name__ == "__main__":
    main()
