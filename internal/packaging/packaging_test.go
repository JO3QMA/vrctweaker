package packaging

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUserReadmeTXT_includesVersionAndLinks(t *testing.T) {
	got := UserReadmeTXT("1.2.3")
	if !strings.Contains(got, "1.2.3") {
		t.Fatalf("missing version in readme")
	}
	if !strings.Contains(got, "getting-started.md") {
		t.Fatalf("missing getting started link")
	}
	if !strings.Contains(got, "vrchat-tweaker.exe") {
		t.Fatalf("missing exe name")
	}
}

func TestFormatChecksumsTXT(t *testing.T) {
	data := []byte("hello")
	wantHash := sha256.Sum256(data)
	wantLine := hex.EncodeToString(wantHash[:]) + "  " + WindowsExeName
	got := FormatChecksumsTXT(map[string][]byte{
		WindowsExeName: data,
	})
	if strings.TrimSpace(got) != wantLine {
		t.Fatalf("got %q want %q", got, wantLine)
	}
}

func TestWindowsZipBaseName(t *testing.T) {
	got, err := WindowsZipBaseName("0.1.0")
	if err != nil {
		t.Fatal(err)
	}
	if got != "vrchat-tweaker-v0.1.0-windows-amd64" {
		t.Fatal(got)
	}
}

func TestWindowsZipBaseName_rejectsPathTraversal(t *testing.T) {
	for _, ver := range []string{"../evil", "1.0/2", "1..0"} {
		if _, err := WindowsZipBaseName(ver); err == nil {
			t.Fatalf("expected error for version %q", ver)
		}
	}
}

func TestReleaseNotesFromChangelog(t *testing.T) {
	const cl = `# Changelog

## [Unreleased]

### Added
- foo

## [0.1.0] - 2026-01-01

### Added
- bar
`
	notes, err := ReleaseNotesFromChangelog([]byte(cl), "0.1.0")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(notes, "bar") {
		t.Fatalf("notes %q", notes)
	}
	if strings.Contains(notes, "foo") {
		t.Fatalf("unexpected unreleased section")
	}
}

func TestReleaseNotesFromChangelog_missingSection(t *testing.T) {
	_, err := ReleaseNotesFromChangelog([]byte("## [0.1.0]\n\n- x\n"), "9.9.9")
	if !errors.Is(err, ErrNoChangelogSection) {
		t.Fatalf("err = %v, want ErrNoChangelogSection", err)
	}
}

func TestReleaseNotesFromChangelog_CRLF(t *testing.T) {
	const cl = "## [0.2.0] - 2026-01-01\r\n\r\n### Added\r\n- crlf-line\r\n"
	notes, err := ReleaseNotesFromChangelog([]byte(cl), "0.2.0")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(notes, "\r") {
		t.Fatalf("unexpected CR in notes: %q", notes)
	}
	if !strings.Contains(notes, "crlf-line") {
		t.Fatalf("notes %q", notes)
	}
}

func TestBuildWindowsReleaseZip_layout(t *testing.T) {
	dir := t.TempDir()
	exePath := filepath.Join(dir, "app.exe")
	if err := os.WriteFile(exePath, []byte("fake-exe"), 0o644); err != nil {
		t.Fatal(err)
	}
	licensePath := filepath.Join(dir, "LICENSE")
	if err := os.WriteFile(licensePath, []byte("MIT"), 0o644); err != nil {
		t.Fatal(err)
	}

	outZip := filepath.Join(dir, "out.zip")
	sum, err := BuildWindowsReleaseZip(BuildWindowsReleaseInput{
		Version:     "9.9.9",
		ExePath:     exePath,
		LicensePath: licensePath,
		OutputZip:   outZip,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sum) != 64 {
		t.Fatalf("zip sha256 len %d", len(sum))
	}

	r, err := zip.OpenReader(outZip)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	names := map[string]bool{}
	for _, f := range r.File {
		names[f.Name] = true
	}
	for _, want := range []string{WindowsExeName, "LICENSE", ReadmeFileName, ChecksumsFileName} {
		if !names[want] {
			t.Fatalf("missing %q in zip, have %v", want, names)
		}
	}
	for _, f := range r.File {
		var wantMode fs.FileMode
		switch f.Name {
		case WindowsExeName:
			wantMode = 0o755
		default:
			wantMode = 0o644
		}
		if f.Mode() != wantMode {
			t.Fatalf("file %q mode %#o want %#o", f.Name, f.Mode(), wantMode)
		}
	}

	var readmeBuf bytes.Buffer
	for _, f := range r.File {
		if f.Name != ReadmeFileName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(&readmeBuf, rc)
		rc.Close()
	}
	if !strings.Contains(readmeBuf.String(), "9.9.9") {
		t.Fatalf("readme %q", readmeBuf.String())
	}
}
