package main

import (
	_ "embed"
	"log"

	"vrchat-tweaker/internal/appversion"
)

//go:embed wails.json
var wailsJSON []byte

func init() {
	if err := appversion.ApplyWailsJSON(wailsJSON); err != nil {
		log.Printf("app version: parse wails.json: %v", err)
	}
}
