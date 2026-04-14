#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Read project name from project.settings
NAME="/tmp/st/myproject"

cd "$PROJECT_ROOT"
echo "[super] building $NAME..."
go build -o "build/$NAME" src/main.go
echo "[super] build complete -> build/$NAME"
