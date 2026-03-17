#!/usr/bin/env python3
import os
import sys
import argparse
from datetime import datetime

def generate_migration(name):
    timestamp = datetime.now().strftime('%Y%m%d%H%M%S')
    base_path = "backend/migrations"
    os.makedirs(base_path, exist_ok=True)

    up_file = os.path.join(base_path, f"{timestamp}_{name}.up.sql")
    down_file = os.path.join(base_path, f"{timestamp}_{name}.down.sql")

    with open(up_file, 'w') as f:
        f.write(f"-- Migration: {name} (UP)\n-- Created at: {datetime.now()}\n\n")
    
    with open(down_file, 'w') as f:
        f.write(f"-- Migration: {name} (DOWN)\n-- Created at: {datetime.now()}\n\n")

    print(f"✅ Created migration files:")
    print(f"  📄 {up_file}")
    print(f"  📄 {down_file}")

def main():
    parser = argparse.ArgumentParser(description='Go Migration Generator')
    parser.add_argument('name', help='Migration name (e.g., create_users_table)')
    args = parser.parse_args()
    
    generate_migration(args.name)

if __name__ == "__main__":
    main()
