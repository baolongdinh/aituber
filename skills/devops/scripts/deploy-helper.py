#!/usr/bin/env python3
import subprocess
import os
import sys
import shutil

def run_step(cmd, name):
    print(f"🚀 Running: {name}...")
    result = subprocess.run(cmd, shell=True)
    if result.returncode != 0:
        print(f"❌ {name} FAILED.")
        sys.exit(1)
    print(f"✅ {name} SUCCESSFUL.\n")

def main():
    print("📦 STARTING DEPLOYMENT BUILD PIPELINE\n")
    
    # 1. Cleanup
    if os.path.exists("dist"):
        shutil.rmtree("dist")
    os.makedirs("dist", exist_ok=True)
    
    # 2. Build Backend
    backend_cmd = "cd backend && CGO_ENABLED=0 GOOS=linux go build -o ../dist/server ./cmd/main.go"
    run_step(backend_cmd, "Backend Compilation (Linux Binary)")
    
    # 3. Build Frontend
    frontend_cmd = "cd frontend && npm install && npm run build"
    run_step(frontend_cmd, "Frontend Production Build")
    
    # 4. Finalize dist package
    if os.path.exists("frontend/dist"):
        shutil.move("frontend/dist", "dist/public")
        print("✅ Moved frontend assets to dist/public\n")
    
    # 5. Copy templates
    if os.path.exists("backend/migrations"):
        shutil.copytree("backend/migrations", "dist/migrations")
        print("✅ Copied migrations to dist/migrations\n")
    
    print("🎉 DEPLOYMENT PACKAGE READY in ./dist/")
    print("Structure:")
    print("  dist/")
    print("    ├── server (binary)")
    print("    ├── public/ (frontend assets)")
    print("    └── migrations/ (SQL migrations)")

if __name__ == "__main__":
    main()
