#!/usr/bin/env bash
# Ensure frontend/wailsjs exists for vue-tsc (gitignored; same stub flow as CI).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WAILJS="$ROOT/frontend/wailsjs"
BINDINGS_DTS="$WAILJS/go/wailsapp/App.d.ts"
MODELS_TS="$WAILJS/go/models.ts"
STALE_MAIN="$WAILJS/go/main"

remove_stale_wails_bindings() {
  if [[ -d "$STALE_MAIN" ]]; then
    rm -rf "$STALE_MAIN"
  fi
}

wails_bindings_ready() {
  [[ -f "$BINDINGS_DTS" && -f "$MODELS_TS" ]]
}

if wails_bindings_ready; then
  remove_stale_wails_bindings
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
remove_stale_wails_bindings
if ! wails_bindings_ready; then
  echo "wails generate module did not produce wailsapp bindings (App.d.ts and models.ts)" >&2
  exit 1
fi
