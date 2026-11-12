#!/usr/bin/env bash
# Playwright 公式 Docker イメージ（CI と同じ Chromium + OS）でビジュアルテストを実行する。
set -euo pipefail

FRONTEND_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "$FRONTEND_ROOT/.." && pwd)"
UPDATE_SNAPSHOTS="${1:-}"

cd "$FRONTEND_ROOT"
if [[ ! -d node_modules/@playwright/test ]]; then
  pnpm install --frozen-lockfile
fi

PLAYWRIGHT_VERSION="$(node -p "require('@playwright/test/package.json').version")"
IMAGE="mcr.microsoft.com/playwright:v${PLAYWRIGHT_VERSION}-noble"

echo "Using Docker image: ${IMAGE}"

docker run --rm \
  -v "${REPO_ROOT}:/work" \
  -w /work/frontend \
  -e CI=1 \
  "${IMAGE}" \
  bash -lc "
    set -euo pipefail
    apt-get update -qq
    apt-get install -y -qq fonts-noto-cjk
    fc-cache -f
    pnpm install --frozen-lockfile
    pnpm exec playwright install chromium
    if [[ \"${UPDATE_SNAPSHOTS}\" == \"--update-snapshots\" ]]; then
      pnpm run test:visual:update
    else
      pnpm run test:visual
    fi
  "
