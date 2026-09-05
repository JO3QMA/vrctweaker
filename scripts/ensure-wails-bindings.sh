#!/usr/bin/env bash
# Ensure frontend/wailsjs exists for vue-tsc (gitignored; same stub flow as CI).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ -f "$ROOT/frontend/wailsjs/go/main/App.d.ts" ]]; then
  exit 0
fi

if ! command -v wails >/dev/null 2>&1; then
  echo "wails CLI not found. Install: go install github.com/wailsapp/wails/v2/cmd/wails@latest" >&2
  echo "CI and devcontainer images include wails; local hooks need it for vue-tsc bindings." >&2
  exit 1
fi

# wails generate module expects a minimal frontend/dist for some project layouts.
mkdir -p "$ROOT/frontend/dist"
echo '<!DOCTYPE html><html></html>' >"$ROOT/frontend/dist/index.html"
(cd "$ROOT" && wails generate module)
rm -f "$ROOT/frontend/dist/index.html"
rmdir "$ROOT/frontend/dist" 2>/dev/null || true
if [[ ! -f "$ROOT/frontend/wailsjs/go/main/App.d.ts" ]]; then
  echo "wails generate module did not produce App.d.ts" >&2
  exit 1
fi
