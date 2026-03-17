#!/usr/bin/env python3
import subprocess
import sys
import json

def run_audit(name, cmd):
    print(f"🛡️  Running {name} audit...")
    try:
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        # Some audit tools exit with non-zero codes if vulnerabilities are found
        return result.stdout, result.stderr
    except Exception as e:
        return "", str(e)

def main():
    print("🚩 Starting Security Audit...\n")
    
    # 1. Go Audit (gosec)
    go_out, go_err = run_audit("Go (gosec)", "cd backend && gosec -fmt json ./...")
    
    # 2. NPM Audit
    npm_out, npm_err = run_audit("NPM (audit)", "cd frontend && npm audit --json")
    
    print("\n" + "="*30)
    print("📊 SECURITY SUMMARY")
    
    # Process Go results
    try:
        if go_out:
            go_data = json.loads(go_out)
            high = go_data.get('Stats', {}).get('high', 0)
            critical = go_data.get('Stats', {}).get('critical', 0)
            print(f"Backend: {critical} Critical, {high} High vulnerabilities found.")
    except:
        print("Backend audit: No gosec data or command failed (ignore if gosec not installed).")

    # Process NPM results
    try:
        if npm_out:
            npm_data = json.loads(npm_out)
            vulns = npm_data.get('metadata', {}).get('vulnerabilities', {})
            critical = vulns.get('critical', 0)
            high = vulns.get('high', 0)
            print(f"Frontend: {critical} Critical, {high} High vulnerabilities found.")
    except:
        print("Frontend audit: No npm audit data or command failed.")

    print("="*30)
    print("\n👉 Use 'gosec' and 'npm audit fix' for details and remediation.")

if __name__ == "__main__":
    main()
