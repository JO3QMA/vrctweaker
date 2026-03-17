#!/usr/bin/env bash
set -euo pipefail

# Use writable workspace for Go module cache (avoids /go permission issues)
export GOMODCACHE="${GOMODCACHE:-$(pwd)/.gomodcache}"
export GOPATH="${GOPATH:-$(pwd)/.go}"
mkdir -p "$GOMODCACHE" "$GOPATH"

if [[ -f "go.mod" ]]; then
  go mod download
fi

if [[ -f "frontend/package.json" ]]; then
  cd frontend && pnpm install && cd ..
fi

