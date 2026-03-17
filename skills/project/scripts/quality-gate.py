#!/usr/bin/env python3
import subprocess
import os
import sys
import argparse

# ANSI Colors
GREEN = "\033[92m"
RED = "\033[91m"
YELLOW = "\033[93m"
CYAN = "\033[96m"
RESET = "\033[0m"

def run_check(cmd, name):
    print(f"{CYAN}🔍 Checking {name}...{RESET}")
    try:
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        if result.returncode == 0:
            print(f"  {GREEN}✅ PASS{RESET}")
            return True
        else:
            print(f"  {RED}❌ FAIL{RESET}")
            print(f"{YELLOW}Output:{RESET}\n{result.stdout}{result.stderr}")
            return False
    except Exception as e:
        print(f"  {RED}💥 ERROR: {str(e)}{RESET}")
        return False

def main():
    parser = argparse.ArgumentParser(description='Unified Quality Gate')
    parser.add_argument('--backend-only', action='store_true', help='Only check backend')
    parser.add_argument('--frontend-only', action='store_true', help='Only check frontend')
    args = parser.parse_args()

    print(f"\n{CYAN}🛡️  ENTERING PROJECT QUALITY GATE{RESET}")
    print("-" * 40)

    overall_pass = True

    # Backend Checks
    if not args.frontend_only and os.path.exists("backend"):
        print(f"\n{CYAN}--- Backend Quality ---{RESET}")
        overall_pass &= run_check("cd backend && go vet ./...", "Go Vet (Linter)")
        overall_pass &= run_check("cd backend && go test ./...", "Go Unit Tests")
    
    # Frontend Checks
    if not args.backend_only and os.path.exists("frontend"):
        print(f"\n{CYAN}--- Frontend Quality ---{RESET}")
        overall_pass &= run_check("cd frontend && npm run lint", "NPM Lint")
        overall_pass &= run_check("cd frontend && npm test -- --run", "Vitest Unit Tests")

    print("\n" + "=" * 40)
    if overall_pass:
        print(f"{GREEN}🎉 ALL QUALITY CHECKS PASSED! READY FOR SIGN-OFF.{RESET}")
    else:
        print(f"{RED}🛑 QUALITY GATE FAILED. Please fix issues before proceeding.{RESET}")
        sys.exit(1)

if __name__ == "__main__":
    main()
