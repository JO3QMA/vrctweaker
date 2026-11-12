#!/usr/bin/env bash
# Playwright公式Dockerイメージ（CIと同じChromium+OS）でビジュアルテストを実行する。
set -euo pipefail

FRONTEND_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REPO_ROOT="$(cd "$FRONTEND_ROOT/.." && pwd)"
UPDATE_SNAPSHOTS="${1:-}"

HOST_UID="$(id -u)"
HOST_GID="$(id -g)"
NODE_MODULES_VOLUME="vrctweaker-visual-node_modules"

PLAYWRIGHT_VERSION="$(
  node -p "require('${FRONTEND_ROOT}/package.json').devDependencies['@playwright/test'].replace(/^\\^/, '')" 2>/dev/null ||
    true
)"
if [[ -z "${PLAYWRIGHT_VERSION}" ]]; then
  PLAYWRIGHT_VERSION="$(
    cd "${FRONTEND_ROOT}" && node -p "require('@playwright/test/package.json').version" 2>/dev/null || echo "1.64.1"
  )"
fi

IMAGE="mcr.microsoft.com/playwright:v${PLAYWRIGHT_VERSION}-noble"
echo "Using Docker image: ${IMAGE}"

DOCKER=(docker)
if ! docker info >/dev/null 2>&1; then
  if sudo docker info >/dev/null 2>&1; then
    DOCKER=(sudo docker)
  fi
fi

chown_visual_artifacts() {
  local maybe_sudo=()
  if [[ "${DOCKER[0]}" == "sudo" ]]; then
    maybe_sudo=(sudo)
  fi
  "${maybe_sudo[@]}" chown -R "${HOST_UID}:${HOST_GID}" \
    "${FRONTEND_ROOT}/e2e/visual" \
    "${FRONTEND_ROOT}/test-results" \
    "${FRONTEND_ROOT}/playwright-report" 2>/dev/null || true
}

VISUAL_CMD="pnpm run test:visual"
if [[ "${UPDATE_SNAPSHOTS}" == "--update-snapshots" ]]; then
  VISUAL_CMD="pnpm run test:visual:update"
fi

"${DOCKER[@]}" run --rm \
  -v "${REPO_ROOT}:/work" \
  -v "${NODE_MODULES_VOLUME}:/work/frontend/node_modules" \
  -w /work/frontend \
  -e CI=1 \
  -e HOST_UID="${HOST_UID}" \
  -e HOST_GID="${HOST_GID}" \
  "${IMAGE}" \
  bash -lc "
    set -euo pipefail
    apt-get update -qq
    apt-get install -y -qq fonts-noto-cjk
    fc-cache -f
    npm install -g pnpm@9
    pnpm install --frozen-lockfile
    pnpm exec playwright install chromium
    ${VISUAL_CMD}
    chown -R \"\${HOST_UID}:\${HOST_GID}\" /work/frontend/e2e/visual || true
    chown -R \"\${HOST_UID}:\${HOST_GID}\" /work/frontend/test-results /work/frontend/playwright-report 2>/dev/null || true
  "

chown_visual_artifacts
