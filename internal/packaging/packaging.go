package packaging

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	WindowsExeName    = "vrchat-tweaker.exe"
	ReadmeFileName    = "README.txt"
	ChecksumsFileName = "checksums.txt"
	gettingStartedURL = "https://github.com/JO3QMA/vrctweaker/blob/main/docs/user/getting-started.md"
	githubReleasesURL = "https://github.com/JO3QMA/vrctweaker/releases"
)

// WindowsZipBaseName returns the archive file base name (without .zip) for a release version.
func WindowsZipBaseName(version string) string {
	v := strings.TrimSpace(version)
	if v == "" {
		v = "unknown"
	}
	return fmt.Sprintf("vrchat-tweaker-v%s-windows-amd64", v)
}

// UserReadmeTXT is the user-facing README shipped inside the Windows zip.
func UserReadmeTXT(version string) string {
	v := strings.TrimSpace(version)
	if v == "" {
		v = "unknown"
	}
	return strings.TrimRight(fmt.Sprintf(`VRChat Tweaker v%s
========================

Windows 向けデスクトップアプリ（非公式）。VRChat 株式会社の製品ではありません。

はじめに（オンライン）:
%s

GitHub Releases:
%s

同梱ファイル:
  %s  - アプリ本体
  LICENSE             - MIT ライセンス
  %s       - 各ファイルの SHA256
  %s        - このファイル

起動: zip を展開し %s を実行してください。
`, v, gettingStartedURL, githubReleasesURL, WindowsExeName, ChecksumsFileName, ReadmeFileName, WindowsExeName), "\n") + "\n"
}

// FormatChecksumsTXT builds checksums.txt content (GNU coreutils sha256sum format).
func FormatChecksumsTXT(files map[string][]byte) string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, name := range names {
		sum := sha256.Sum256(files[name])
		b.WriteString(hex.EncodeToString(sum[:]))
		b.WriteString("  ")
		b.WriteString(name)
		b.WriteByte('\n')
	}
	return b.String()
}

// BuildWindowsReleaseInput describes inputs for BuildWindowsReleaseZip.
type BuildWindowsReleaseInput struct {
	Version     string
	ExePath     string
	LicensePath string
	OutputZip   string
}

// BuildWindowsReleaseZip stages README.txt and checksums.txt, then writes the release zip.
// It returns the SHA256 hex digest of the zip file.
func BuildWindowsReleaseZip(in BuildWindowsReleaseInput) (zipSHA256 string, err error) {
	if strings.TrimSpace(in.Version) == "" {
		return "", fmt.Errorf("packaging: version is required")
	}
	exeData, err := os.ReadFile(in.ExePath)
	if err != nil {
		return "", fmt.Errorf("packaging: read exe: %w", err)
	}
	licenseData, err := os.ReadFile(in.LicensePath)
	if err != nil {
		return "", fmt.Errorf("packaging: read license: %w", err)
	}

	readme := UserReadmeTXT(in.Version)
	checksums := FormatChecksumsTXT(map[string][]byte{
		WindowsExeName: exeData,
		"LICENSE":      licenseData,
		ReadmeFileName: []byte(readme),
	})

	if mkdirErr := os.MkdirAll(filepath.Dir(in.OutputZip), 0o755); mkdirErr != nil {
		return "", mkdirErr
	}
	tmpZip := in.OutputZip + ".tmp"
	if writeErr := writeZip(tmpZip, map[string][]byte{
		WindowsExeName:    exeData,
		"LICENSE":         licenseData,
		ReadmeFileName:    []byte(readme),
		ChecksumsFileName: []byte(checksums),
	}); writeErr != nil {
		return "", writeErr
	}
	zipBytes, readErr := os.ReadFile(tmpZip)
	if readErr != nil {
		_ = os.Remove(tmpZip)
		return "", readErr
	}
	sum := sha256.Sum256(zipBytes)
	zipSHA256 = hex.EncodeToString(sum[:])
	if renameErr := os.Rename(tmpZip, in.OutputZip); renameErr != nil {
		_ = os.Remove(tmpZip)
		return "", renameErr
	}
	return zipSHA256, nil
}

func writeZip(path string, files map[string][]byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	zw := zip.NewWriter(f)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		w, createErr := zw.Create(name)
		if createErr != nil {
			_ = zw.Close()
			_ = os.Remove(path)
			return createErr
		}
		if _, writeErr := w.Write(files[name]); writeErr != nil {
			_ = zw.Close()
			_ = os.Remove(path)
			return writeErr
		}
	}
	if closeErr := zw.Close(); closeErr != nil {
		_ = os.Remove(path)
		return closeErr
	}
	return nil
}

// FileSHA256Hex returns the SHA256 hex digest of a file.
func FileSHA256Hex(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
