#!/usr/bin/env bash
# Decode staged base64 chunks into exact VT app icons. DO NOT regenerate/redraw icons.
set -euo pipefail
cd "$(dirname "$0")"
PNG_OUT="../appicon.png"
ICO_OUT="../windows/icon.ico"
mkdir -p "$(dirname "$ICO_OUT")"
cat $(ls appicon.png.b64.[0-9]* | sort) | tr -d '\n\r\t ' | base64 -d > "$PNG_OUT"
cat $(ls icon.ico.b64.[0-9]* | sort) | tr -d '\n\r\t ' | base64 -d > "$ICO_OUT"
PNG_SIZE=$(wc -c < "$PNG_OUT" | tr -d ' ')
ICO_SIZE=$(wc -c < "$ICO_OUT" | tr -d ' ')
echo "appicon.png size=$PNG_SIZE"
echo "icon.ico size=$ICO_SIZE"
test "$PNG_SIZE" = "303021" || { echo "FAIL: expected png 303021 got $PNG_SIZE"; exit 1; }
test "$ICO_SIZE" = "67850" || { echo "FAIL: expected ico 67850 got $ICO_SIZE"; exit 1; }
if command -v sha256sum >/dev/null 2>&1; then
  PNG_SHA=$(sha256sum "$PNG_OUT" | awk '{print $1}')
  ICO_SHA=$(sha256sum "$ICO_OUT" | awk '{print $1}')
  echo "png sha256=$PNG_SHA"
  echo "ico sha256=$ICO_SHA"
  test "$PNG_SHA" = "511a47b01ceb24e0dfd100dedb1c6333e97391b84f6e428e7662bd59c02c7b2d" || { echo "FAIL: png sha mismatch"; exit 1; }
  test "$ICO_SHA" = "26ee79745390f7015738faa20b1189c96f63e3a6f73975e6837acba29ed75537" || { echo "FAIL: ico sha mismatch"; exit 1; }
fi
STAGING_DIR="$(pwd)"
cd ..
rm -rf "$STAGING_DIR"
echo "OK: icons written and staging removed"
