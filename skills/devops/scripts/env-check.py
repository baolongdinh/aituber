#!/usr/bin/env python3
import subprocess
import os
import sys
import re

# Recommended Minimum Versions
MIN_GO_VERSION = "1.22"
MIN_NODE_VERSION = "20"

def get_installed_version(cmd):
    try:
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        if result.returncode == 0:
            return result.stdout.strip()
    except Exception:
        pass
    return None

def version_ge(v1, v2):
    """Simple version comparison (major.minor)"""
    def parse(v):
        match = re.search(r"(\d+\.\d+)", v)
        return [int(x) for x in match.group(1).split(".")] if match else [0, 0]
    return parse(v1) >= parse(v2)

def check_runtime(cmd, name, min_ver):
    print(f"🔍 Checking {name} (Min required: {min_ver})...")
    v = get_installed_version(cmd)
    if v:
        if version_ge(v, min_ver):
            print(f"  ✅ OK: {v}")
            return True
        else:
            print(f"  ⚠️  WARNING: {v} is below recommended {min_ver}.")
            return True # Still allow but warn
    print(f"  ❌ {name} NOT FOUND.")
    return False

def check_env_file(path, required_vars):
    print(f"🔍 Checking {path}...")
    if not os.path.exists(path):
        print(f"  💡 Suggestion: Create {path} if needed.")
        return True
    
    with open(path, 'r') as f:
        content = f.read()
        missing = [v for v in required_vars if v not in content]
        if not missing:
            print(f"  ✅ OK: All required variables present.")
            return True
        else:
            print(f"  ⚠️  WARNING: Missing potential vars: {', '.join(missing)}")
            return True

def main():
    print("🚀 ADVANCED ENVIRONMENT CHECK (Wave 5)\n")
    
    success = True
    success &= check_runtime("go version", "Go", MIN_GO_VERSION)
    success &= check_runtime("node -v", "Node.js", MIN_NODE_VERSION)
    
    # Optional tools
    check_runtime("docker -v", "Docker", "20.10")
    
    print("\n" + "-"*30)
    
    if os.path.exists("backend"):
        check_env_file("backend/.env.example", ["POSTGRES_URL", "JWT_SECRET"])
    if os.path.exists("frontend"):
        check_env_file("frontend/.env.example", ["VITE_API_URL"])
    
    print("\n" + "-"*30)
    if success:
        print("✨ Environment check finished. You're ready to build!")
    else:
        print("🛑 Critical runtimes missing. Please install Go/Node.")

if __name__ == "__main__":
    main()
