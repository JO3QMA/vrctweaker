#!/usr/bin/env bash
# WSL 上の var/ から Windows 側の VRChat Tweaker DB・VRChat ログへ symlink を張る。
# リポジトリルートで実行: make link-var
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

LOCAL_ENV="$ROOT/var/local.env"
if [[ -f "$LOCAL_ENV" ]]; then
  # shellcheck disable=SC1090
  source "$LOCAL_ENV"
fi

WIN_USER="${VRCTWEAKER_WIN_USER:-}"
if [[ -z "$WIN_USER" ]]; then
  echo "error: VRCTWEAKER_WIN_USER が未設定です。" >&2
  echo "  cp var/local.env.example var/local.env" >&2
  echo "  # var/local.env の Windows ユーザー名を編集してから再実行" >&2
  exit 1
fi

WIN_DATA="${VRCTWEAKER_WIN_DATA_DIR:-/mnt/c/Users/${WIN_USER}/AppData/Roaming/vrchat-tweaker}"
WIN_LOGS="${VRCTWEAKER_WIN_VRC_LOG_DIR:-/mnt/c/Users/${WIN_USER}/AppData/LocalLow/VRChat/VRChat}"

if [[ ! -d /mnt/c ]]; then
  echo "error: /mnt/c がありません。WSL で Windows ドライブがマウントされている環境で実行してください。" >&2
  exit 1
fi

link_dir() {
  local label="$1" target="$2" link="$3"
  if [[ ! -e "$target" ]]; then
    echo "warning: skip ${label} — 参照先がありません: ${target}" >&2
    return 0
  fi
  mkdir -p "$(dirname "$link")"
  ln -sfn "$target" "$link"
  echo "linked ${link} -> ${target}"
}

mkdir -p var/data var/logs

link_dir "VRChat Tweaker data" "$WIN_DATA" "var/data/win"
link_dir "VRChat logs" "$WIN_LOGS" "var/logs/vrchat"

latest="$(ls -t var/logs/vrchat/output_log_*.txt 2>/dev/null | head -1 || true)"
if [[ -n "$latest" ]]; then
  ln -sfn "vrchat/$(basename "$latest")" "var/logs/latest-output_log.txt"
  echo "linked var/logs/latest-output_log.txt -> vrchat/$(basename "$latest")"
fi

echo ""
echo "Agent / 手元からの参照例:"
echo "  var/data/win/vrchat-tweaker.db"
echo "  var/logs/latest-output_log.txt"
echo "  var/logs/vrchat/"
