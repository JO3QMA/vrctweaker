#!/usr/bin/env bash
# Decode staged base64 chunks into exact VT app icons. DO NOT regenerate/redraw icons.
set -euo pipefail
cd "$(dirname "$0")"
PNG_OUT="../appicon.png"
ICO_OUT="../windows/icon.ico"
mkdir -p "$(dirname "$ICO_OUT")"

shopt -s nullglob
PNG_PARTS=( $(ls appicon.png.b64.[0-9][0-9][0-9] | sort) )
echo "Found ${#PNG_PARTS[@]} png staging parts"
test "${#PNG_PARTS[@]}" = "51" || { echo "FAIL: expected 51 png parts (000-050), got ${#PNG_PARTS[@]}"; ls -la appicon.png.b64.* || true; exit 1; }

cat "${PNG_PARTS[@]}" | tr -d '\n\r\t ' | base64 -d > "$PNG_OUT"
PNG_SIZE=$(wc -c < "$PNG_OUT" | tr -d ' ')
echo "appicon.png size=$PNG_SIZE"
test "$PNG_SIZE" = "303021" || { echo "FAIL: expected png 303021 got $PNG_SIZE"; exit 1; }
python3 - <<'PY'
p=open('../appicon.png','rb').read(8)
assert p==b'\x89PNG\r\n\x1a\n', p
print('PNG magic OK')
PY

if command -v sha256sum >/dev/null 2>&1; then
  PNG_SHA=$(sha256sum "$PNG_OUT" | awk '{print $1}')
  echo "png sha256=$PNG_SHA"
  test "$PNG_SHA" = "511a47b01ceb24e0dfd100dedb1c6333e97391b84f6e428e7662bd59c02c7b2d" || { echo "FAIL: png sha mismatch"; exit 1; }
fi

if command -v convert >/dev/null 2>&1; then
  convert "$PNG_OUT" -define icon:auto-resize=256,128,64,48,32,16 "$ICO_OUT"
elif python3 -c 'from PIL import Image' 2>/dev/null; then
  python3 - <<'PY'
from PIL import Image
im=Image.open('../appicon.png').convert('RGBA')
sizes=[(16,16),(32,32),(48,48),(64,64),(128,128),(256,256)]
im.save('../windows/icon.ico', format='ICO', sizes=sizes)
print('Wrote ICO via Pillow')
PY
else
  echo "FAIL: need ImageMagick convert or Pillow"; exit 1
fi
ICO_SIZE=$(wc -c < "$ICO_OUT" | tr -d ' ')
echo "icon.ico size=$ICO_SIZE"
test "$ICO_SIZE" -gt 1000 || { echo "FAIL: ico too small"; exit 1; }
echo "OK: build/appicon.png and build/windows/icon.ico written from approved staging b64"
