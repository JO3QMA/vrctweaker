#!/usr/bin/env bash
# Ensure frontend/wailsjs exists for vue-tsc (gitignored; same stub flow as CI).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ -f "$ROOT/frontend/wailsjs/go/main/App.d.ts" ]]; then
  exit 0
fi

if ! command -v wails >/dev/null 2>&1; then
  echo "wails CLI not found; install Wails v2 or run: go install github.com/wailsapp/wails/v2/cmd/wails@latest" >&2
  exit 1
fi

mkdir -p "$ROOT/frontend/dist"
echo '<!DOCTYPE html><html></html>' >"$ROOT/frontend/dist/index.html"
(cd "$ROOT" && wails generate module)
if [[ ! -f "$ROOT/frontend/wailsjs/go/main/App.d.ts" ]]; then
  echo "wails generate module did not produce App.d.ts" >&2
  exit 1
fi
