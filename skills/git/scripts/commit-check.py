#!/usr/bin/env python3
import sys
import re
import subprocess

def check_commit_message(message):
    # Conventional Commits regex: type(scope)!: description
    pattern = r'^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(?:\(.+\))?!?: .+$'
    if not re.match(pattern, message):
        return False, "Commit message does not follow Conventional Commits format (e.g., 'feat: add something')"
    return True, ""

def get_staged_files():
    try:
        result = subprocess.run(['git', 'diff', '--cached', '--name-only'], capture_output=True, text=True, check=True)
        return result.stdout.splitlines()
    except subprocess.CalledProcessError:
        return []

def check_trailing_whitespace(files):
    errors = []
    for file in files:
        try:
            with open(file, 'r', errors='ignore') as f:
                for i, line in enumerate(f, 1):
                    if line.rstrip('\r\n') != line.rstrip():
                        errors.append(f"{file}:{i} - Trailing whitespace found")
        except Exception:
            continue
    return errors

def main():
    if len(sys.argv) < 2:
        print("Usage: python3 commit-check.py \"commit message\"")
        sys.exit(1)

    message = sys.argv[1]
    
    # 1. Check commit message format
    is_valid, error_msg = check_commit_message(message)
    if not is_valid:
        print(f"❌ {error_msg}")
        sys.exit(1)
    
    # 2. Check for trailing whitespace in staged files
    staged_files = get_staged_files()
    if staged_files:
        whitespace_errors = check_trailing_whitespace(staged_files)
        if whitespace_errors:
            print("❌ Quality check failed:")
            for err in whitespace_errors:
                print(f"  {err}")
            sys.exit(1)

    print("✅ Commit check passed!")

if __name__ == "__main__":
    main()
