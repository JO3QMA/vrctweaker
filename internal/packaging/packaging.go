package packaging

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
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

func formatChecksumsFromHex(nameToHash map[string]string) string {
	names := make([]string, 0, len(nameToHash))
	for name := range nameToHash {
		names = append(names, name)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, name := range names {
		b.WriteString(nameToHash[name])
		b.WriteString("  ")
		b.WriteString(name)
		b.WriteByte('\n')
	}
	return b.String()
}

func sha256HexFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()
	h := sha256.New()
	if _, copyErr := io.Copy(h, f); copyErr != nil {
		return "", copyErr
	}
	return hex.EncodeToString(h.Sum(nil)), nil
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
	exeHash, err := sha256HexFile(in.ExePath)
	if err != nil {
		return "", fmt.Errorf("packaging: hash exe: %w", err)
	}
	licenseHash, err := sha256HexFile(in.LicensePath)
	if err != nil {
		return "", fmt.Errorf("packaging: hash license: %w", err)
	}

	readme := UserReadmeTXT(in.Version)
	readmeSum := sha256.Sum256([]byte(readme))
	readmeHash := hex.EncodeToString(readmeSum[:])
	checksums := formatChecksumsFromHex(map[string]string{
		WindowsExeName: exeHash,
		"LICENSE":      licenseHash,
		ReadmeFileName: readmeHash,
	})

	if mkdirErr := os.MkdirAll(filepath.Dir(in.OutputZip), 0o755); mkdirErr != nil {
		return "", mkdirErr
	}
	tmpZip := in.OutputZip + ".tmp"
	zipSHA256, writeErr := writeReleaseZip(tmpZip, []releaseZipEntry{
		{name: WindowsExeName, path: in.ExePath},
		{name: "LICENSE", path: in.LicensePath},
		{name: ReadmeFileName, data: []byte(readme)},
		{name: ChecksumsFileName, data: []byte(checksums)},
	})
	if writeErr != nil {
		return "", writeErr
	}
	if renameErr := os.Rename(tmpZip, in.OutputZip); renameErr != nil {
		_ = os.Remove(tmpZip)
		return "", renameErr
	}
	return zipSHA256, nil
}

type releaseZipEntry struct {
	name string
	path string
	data []byte
}

func writeReleaseZip(path string, entries []releaseZipEntry) (zipSHA256 string, err error) {
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()

	zipHasher := sha256.New()
	mw := io.MultiWriter(f, zipHasher)
	zw := zip.NewWriter(mw)

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.name)
	}
	sort.Strings(names)

	for _, name := range names {
		var entry releaseZipEntry
		for _, e := range entries {
			if e.name == name {
				entry = e
				break
			}
		}
		w, createErr := zw.Create(name)
		if createErr != nil {
			_ = zw.Close()
			_ = os.Remove(path)
			return "", createErr
		}
		if entry.data != nil {
			if _, writeErr := w.Write(entry.data); writeErr != nil {
				_ = zw.Close()
				_ = os.Remove(path)
				return "", writeErr
			}
			continue
		}
		src, openErr := os.Open(entry.path)
		if openErr != nil {
			_ = zw.Close()
			_ = os.Remove(path)
			return "", openErr
		}
		_, copyErr := io.Copy(w, src)
		closeErr := src.Close()
		if copyErr != nil {
			_ = zw.Close()
			_ = os.Remove(path)
			return "", copyErr
		}
		if closeErr != nil {
			_ = zw.Close()
			_ = os.Remove(path)
			return "", closeErr
		}
	}
	if closeErr := zw.Close(); closeErr != nil {
		_ = os.Remove(path)
		return "", closeErr
	}
	if syncErr := f.Sync(); syncErr != nil {
		_ = os.Remove(path)
		return "", syncErr
	}
	return hex.EncodeToString(zipHasher.Sum(nil)), nil
}
