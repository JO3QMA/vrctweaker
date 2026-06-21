#!/usr/bin/env bash
# E2E テスト（Playwright）環境の初回セットアップ。
# リポジトリルートで実行: bash scripts/setup-e2e.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if ! command -v pnpm >/dev/null 2>&1; then
  echo "error: pnpm が見つかりません。Node.js 20 系と pnpm をインストールしてください。" >&2
  exit 1
fi

echo "==> frontend 依存関係をインストール"
make install-front

echo "==> Playwright chromium（システム依存含む）をインストール"
make test-e2e-install

echo "==> E2E スモークテスト"
make test-e2e

echo "==> 完了。以降は make test-e2e で E2E を実行できます。"
