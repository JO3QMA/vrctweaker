# Approved app icon (base64 chunks)

Temporary transfer of the approved 1024x1024 PNG.

Assemble:
```bash
cd vrctweaker-icon/appicon.b64
cat $(printf '%02d.txt\n' $(seq 0 $(($(cat n.txt)-1)))) | base64 -d > ../../build/appicon.png
```

Expected decoded size ~303021 bytes. Do not regenerate artwork.
