#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
# Assemble appicon.png from base64 parts
cat $(ls -1 appicon.b64.* | sort -V) | tr -d '\n\r ' | base64 -d > ../appicon.png
# Assemble windows icon
mkdir -p ../windows
cat $(ls -1 icon.b64.* | sort -V) | tr -d '\n\r ' | base64 -d > ../windows/icon.ico
PNG_SIZE=$(wc -c < ../appicon.png | tr -d ' ')
ICO_SIZE=$(wc -c < ../windows/icon.ico | tr -d ' ')
PNG_SHA=$(sha256sum ../appicon.png | awk '{print $1}')
ICO_SHA=$(sha256sum ../windows/icon.ico | awk '{print $1}')
echo "appicon.png size=$PNG_SIZE sha256=$PNG_SHA"
echo "icon.ico size=$ICO_SIZE sha256=$ICO_SHA"
test "$PNG_SIZE" = "303021"
test "$ICO_SIZE" = "67850"
test "$PNG_SHA" = "511a47b01ceb24e0dfd100dedb1c6333e97391b84f6e428e7662bd59c02c7b2d"
test "$ICO_SHA" = "26ee79745390f7015738faa20b1189c96f63e3a6f73975e6837acba29ed75537"
# remove staging after success
cd ..
rm -rf _icon_staging
echo "STAGING_REMOVED OK"
