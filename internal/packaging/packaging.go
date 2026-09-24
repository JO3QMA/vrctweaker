package packaging

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"vrchat-tweaker/internal/appversion"
)

const (
	WindowsExeName    = "vrchat-tweaker.exe"
	ReadmeFileName    = "README.txt"
	ChecksumsFileName = "checksums.txt"
	gettingStartedURL = "https://github.com/JO3QMA/vrctweaker/blob/main/docs/user/getting-started.md"
	githubReleasesURL = "https://github.com/JO3QMA/vrctweaker/releases"
)

// WindowsZipBaseName returns the archive file base name (without .zip) for a release version.
func WindowsZipBaseName(version string) (string, error) {
	if err := appversion.ValidateProductVersion(version); err != nil {
		return "", err
	}
	v := strings.TrimSpace(version)
	return fmt.Sprintf("vrchat-tweaker-v%s-windows-amd64", v), nil
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
	if err = appversion.ValidateProductVersion(in.Version); err != nil {
		return "", err
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
	zipSHA256, writeErr := writeReleaseZip(tmpZip, []releaseZipEntry{
		{name: WindowsExeName, data: exeData},
		{name: "LICENSE", data: licenseData},
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
	data []byte
}

func zipEntryFileMode(name string) fs.FileMode {
	if strings.EqualFold(name, WindowsExeName) || strings.HasSuffix(strings.ToLower(name), ".exe") {
		return 0o755
	}
	return 0o644
}

func writeReleaseZip(path string, entries []releaseZipEntry) (zipSHA256 string, err error) {
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	keepFile := false
	defer func() {
		closeErr := f.Close()
		if closeErr != nil {
			if err == nil {
				err = closeErr
			}
		}
		if !keepFile {
			_ = os.Remove(path)
		}
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
		hdr := &zip.FileHeader{
			Name:   name,
			Method: zip.Deflate,
		}
		hdr.SetMode(zipEntryFileMode(name))
		w, createErr := zw.CreateHeader(hdr)
		if createErr != nil {
			return "", createErr
		}
		if _, writeErr := w.Write(entry.data); writeErr != nil {
			return "", writeErr
		}
	}
	if closeErr := zw.Close(); closeErr != nil {
		return "", closeErr
	}
	if syncErr := f.Sync(); syncErr != nil {
		return "", syncErr
	}
	keepFile = true
	return hex.EncodeToString(zipHasher.Sum(nil)), nil
}
