#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# Disable Corepack download prompt (non-interactive)
export COREPACK_ENABLE_DOWNLOAD_PROMPT=0

# Use writable workspace for Go module cache (avoids /go permission issues)
export GOMODCACHE="${GOMODCACHE:-$REPO_ROOT/.gomodcache}"
export GOPATH="${GOPATH:-$REPO_ROOT/.go}"
mkdir -p "$GOMODCACHE" "$GOPATH"
export PATH="$GOPATH/bin:$PATH"

cd "$REPO_ROOT"

if [[ -f "go.mod" ]]; then
  go mod download
fi

if [[ -f "frontend/package.json" ]]; then
  cd frontend && pnpm install && pnpm exec playwright install --with-deps chromium && cd "$REPO_ROOT"
fi

if [[ -f "$REPO_ROOT/lefthook.yml" ]] && command -v lefthook >/dev/null 2>&1; then
  lefthook install -f "$REPO_ROOT/lefthook.yml"
fi
