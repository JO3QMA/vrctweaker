package desktop

import (
	"path/filepath"
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
