#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

NAME="super"
SETTINGS="$PROJECT_ROOT/project.settings"
MAIN_GO="$PROJECT_ROOT/src/main.go"

# Auto-increment patch version from project.settings
CURRENT=$(grep -E '^\s+version = ' "$SETTINGS" 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
if [ -n "$CURRENT" ]; then
  MAJOR=$(echo "$CURRENT" | cut -d. -f1)
  MINOR=$(echo "$CURRENT" | cut -d. -f2)
  PATCH=$(echo "$CURRENT" | cut -d. -f3)
  NEW_VERSION="$MAJOR.$MINOR.$((PATCH + 1))"
  sed -i.bak "s/^\(  version = \"\)[0-9]*\.[0-9]*\.[0-9]*/\1$NEW_VERSION/" "$SETTINGS" && rm -f "$SETTINGS.bak"
  sed -i.bak "s/var version = \"[^\"]*\"/var version = \"$NEW_VERSION\"/" "$MAIN_GO" && rm -f "$MAIN_GO.bak"
  echo "[super] version: $CURRENT -> $NEW_VERSION"
else
  NEW_VERSION="dev"
fi

cd "$PROJECT_ROOT"
echo "[super] building $NAME @ v$NEW_VERSION..."
go build -o "build/$NAME" src/*.go
echo "[super] build complete -> build/$NAME"
