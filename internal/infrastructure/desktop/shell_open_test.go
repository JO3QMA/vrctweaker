package desktop

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestOpenFileWithDefaultApp_NotFound(t *testing.T) {
	t.Parallel()
	err := OpenFileWithDefaultApp(filepath.Join(t.TempDir(), "missing.bin"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestRevealInFileManager_NotFound(t *testing.T) {
	t.Parallel()
	err := RevealInFileManager(filepath.Join(t.TempDir(), "missing.bin"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestOpenFolderInFileManager_NotFound(t *testing.T) {
	t.Parallel()
	err := OpenFolderInFileManager(filepath.Join(t.TempDir(), "missing-dir"))
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}

func TestRevealWindows_argFormat(t *testing.T) {
	t.Parallel()
	const abs = `C:\Users\test\file.png`
	got := revealWindows(abs)
	want := "/select," + abs
	if got != want {
		t.Fatalf("revealWindows() = %q, want %q", got, want)
	}
}

func TestOpenFileWindows_nonWindowsStub(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub is not used on windows")
	}
	t.Parallel()
	err := openFileWindows("/tmp/file.txt")
	if err == nil {
		t.Fatal("expected error from non-windows stub")
	}
	if !strings.Contains(err.Error(), "windows-only") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func installFakeBinScript(t *testing.T, name, body string) string {
	t.Helper()
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(binDir, name)
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return binDir
}

func installFakeXDGOpen(t *testing.T) string {
	t.Helper()
	return installFakeBinScript(t, "xdg-open", "#!/bin/sh\nexit 0\n")
}

func installFailingXDGOpen(t *testing.T) string {
	t.Helper()
	return installFakeBinScript(t, "xdg-open", "#!/bin/sh\nexit 1\n")
}

func installFakeNautilus(t *testing.T) string {
	t.Helper()
	return installFakeBinScript(t, "nautilus", "#!/bin/sh\nexit 0\n")
}

func installFailingNautilus(t *testing.T) string {
	t.Helper()
	return installFakeBinScript(t, "nautilus", "#!/bin/sh\nexit 1\n")
}

func TestRevealWindowsExec_fakeExplorer(t *testing.T) {
	installFakeBinScript(t, "explorer", "#!/bin/sh\nexit 0\n")
	err := revealWindowsExec(`C:\Users\test\photo.png`)
	if err != nil {
		t.Fatalf("revealWindowsExec: %v", err)
	}
}
