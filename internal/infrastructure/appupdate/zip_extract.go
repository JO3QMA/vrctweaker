package appupdate

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"vrchat-tweaker/internal/packaging"
)

// ExtractWindowsExe extracts vrchat-tweaker.exe from a release zip into destDir.
func ExtractWindowsExe(zipPath, destDir string) (exePath string, err error) {
	zipPath = strings.TrimSpace(zipPath)
	if zipPath == "" {
		return "", fmt.Errorf("appupdate: zip path is empty")
	}
	if mkdirErr := os.MkdirAll(destDir, 0o700); mkdirErr != nil {
		return "", mkdirErr
	}
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("appupdate: open zip: %w", err)
	}
	defer func() { _ = r.Close() }()

	var found *zip.File
	for _, f := range r.File {
		if filepath.Base(f.Name) == packaging.WindowsExeName {
			found = f
			break
		}
	}
	if found == nil {
		return "", fmt.Errorf("appupdate: %q not in zip", packaging.WindowsExeName)
	}
	destPath := filepath.Join(destDir, packaging.WindowsExeName)
	if err := extractZipFile(found, destPath); err != nil {
		return "", err
	}
	return destPath, nil
}

func extractZipFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("appupdate: open zip entry: %w", err)
	}
	defer func() { _ = rc.Close() }()

	tmp, err := os.CreateTemp(filepath.Dir(destPath), ".exe-*.partial")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := io.Copy(tmp, rc); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("appupdate: extract: %w", err)
	}
	if err := tmp.Chmod(0o755); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("appupdate: replace exe: %w", err)
	}
	return nil
}
