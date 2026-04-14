#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

NAME="/tmp/super-test/nonempty"
BINARY="$PROJECT_ROOT/build/$NAME"

if [ ! -f "$BINARY" ]; then
  echo "[super] binary not found, run build first."
  exit 1
fi

exec "$BINARY" "$@"
