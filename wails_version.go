package main

import (
	_ "embed"
	"log"

	"vrchat-tweaker/internal/appversion"
)

//go:embed wails.json
var wailsJSON []byte

func init() {
	// Product version is embedded at compile time; a parse error is deterministic for
	// every run of that binary. We do not os.Exit here: this is a GUI app and a bad
	// embed should not brick the window. On failure, appversion.Version stays "dev"
	// (see internal/appversion) and settings / User-Agent reflect that until fixed.
	// Startup logs are the release observability hook — search for "app version:".
	if err := appversion.ApplyWailsJSON(wailsJSON); err != nil {
		log.Printf("app version: parse wails.json failed (Version remains %q): %v", appversion.Version, err)
	}
}
