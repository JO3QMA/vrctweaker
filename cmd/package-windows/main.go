package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"vrchat-tweaker/internal/appversion"
	"vrchat-tweaker/internal/packaging"
)

func main() {
	root := flag.String("root", ".", "repository root")
	exe := flag.String("exe", "build/bin/vrchat-tweaker.exe", "path to Windows executable")
	license := flag.String("license", "LICENSE", "path to LICENSE file")
	outDir := flag.String("out", "dist", "output directory for zip")
	flag.Parse()

	wailsPath := filepath.Join(*root, "wails.json")
	wailsData, err := os.ReadFile(wailsPath)
	if err != nil {
		fail("read wails.json: %v", err)
	}
	version, err := appversion.ParseProductVersionFromWailsJSON(wailsData)
	if err != nil {
		fail("parse wails.json: %v", err)
	}
	if version == "" {
		fail("wails.json info.productVersion is empty")
	}
	if err = appversion.ValidateProductVersion(version); err != nil {
		fail("%v", err)
	}

	exePath := *exe
	if !filepath.IsAbs(exePath) {
		exePath = filepath.Join(*root, exePath)
	}
	licensePath := *license
	if !filepath.IsAbs(licensePath) {
		licensePath = filepath.Join(*root, licensePath)
	}
	out := *outDir
	if !filepath.IsAbs(out) {
		out = filepath.Join(*root, out)
	}

	baseName, err := packaging.WindowsZipBaseName(version)
	if err != nil {
		fail("%v", err)
	}
	zipName := baseName + ".zip"
	zipPath := filepath.Join(out, zipName)
	zipSHA, err := packaging.BuildWindowsReleaseZip(packaging.BuildWindowsReleaseInput{
		Version:     version,
		ExePath:     exePath,
		LicensePath: licensePath,
		OutputZip:   zipPath,
	})
	if err != nil {
		fail("%v", err)
	}

	fmt.Printf("version=%s\n", version)
	fmt.Printf("zip=%s\n", zipPath)
	fmt.Printf("zip_sha256=%s\n", zipSHA)
}

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "package-windows: "+format+"\n", args...)
	os.Exit(1)
}
