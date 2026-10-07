#!/usr/bin/env python3
"""Build an Expo/EAS app.

Usage: build_eas            eas build -p android
       build_eas ios        eas build -p ios
       build_eas manual [st]  local gradle assembleRelease in ./android (st = --stacktrace)
"""
import os
import shutil
import subprocess
import sys


def run(cmd, cwd=None):
    exe = shutil.which(cmd[0])
    if not exe:
        print(f"{cmd[0]} not found in PATH")
        sys.exit(127)
    sys.exit(subprocess.run([exe, *cmd[1:]], cwd=cwd).returncode)


def main():
    args = sys.argv[1:]
    if args and args[0] == "manual":
        gradlew = os.path.join("android", "gradlew.bat" if os.name == "nt" else "gradlew")
        if not os.path.isfile(gradlew):
            print(f"{gradlew} not found in {os.getcwd()}")
            sys.exit(1)
        cmd = [os.path.abspath(gradlew), "assembleRelease"]
        if len(args) > 1 and args[1] == "st":
            cmd.append("--stacktrace")
        run(cmd, cwd="android")
    run(["eas", "build", "-p", "ios" if args and args[0] == "ios" else "android"])


if __name__ == "__main__":
    main()
