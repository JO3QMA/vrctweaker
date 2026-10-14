package appupdate

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"vrchat-tweaker/internal/packaging"
)

func TestExtractWindowsExe(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "release.zip")
	f, createErr := os.Create(zipPath)
	if createErr != nil {
		t.Fatal(createErr)
	}
	zw := zip.NewWriter(f)
	w, entryErr := zw.Create(packaging.WindowsExeName)
	if entryErr != nil {
		t.Fatal(entryErr)
	}
	if _, writeErr := w.Write([]byte("MZ-fake")); writeErr != nil {
		t.Fatal(writeErr)
	}
	if closeZwErr := zw.Close(); closeZwErr != nil {
		t.Fatal(closeZwErr)
	}
	if closeErr := f.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}

	outDir := filepath.Join(dir, "staging")
	exe, err := ExtractWindowsExe(zipPath, outDir)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(exe) != packaging.WindowsExeName {
		t.Fatalf("exe path %q", exe)
	}
}
