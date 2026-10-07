#!/usr/bin/env python3
"""Create a branch, commit, push it, then merge it into a target branch.

Usage: fullmerge <new branch> <target branch> <commit message...>
       fullmerge -s     show git status
"""
import subprocess
import sys


def git(*args):
    print()
    r = subprocess.run(["git", *args])
    if r.returncode != 0:
        sys.exit(r.returncode)


def main():
    args = sys.argv[1:]
    if args and args[0] == "-s":
        git("status")
        print("\nuse fullmerge with: new branchname, target branch and then commit message\n")
        return
    if len(args) < 3:
        print(__doc__)
        sys.exit(1)
    new, target, message = args[0], args[1], " ".join(args[2:])

    print(f'"{new} -> {target}"')
    print(f'"{message}"')
    git("pull")
    git("add", ".")
    git("checkout", "-b", new)
    git("commit", "-m", message)
    git("push", "--set-upstream", "origin", new)
    git("checkout", target)
    git("merge", new)


if __name__ == "__main__":
    main()
