package wailsapp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDataDirExistsAsFolder_requiresExistingPath(t *testing.T) {
	dir := t.TempDir()
	if !dataDirExistsAsFolder(dir) {
		t.Fatal("expected existing dir")
	}
	missing := filepath.Join(dir, "missing-sub")
	if dataDirExistsAsFolder(missing) {
		t.Fatal("expected false for missing path")
	}
	parentOnly := filepath.Join(dir, "no-such-child")
	_ = os.MkdirAll(filepath.Dir(parentOnly), 0o700)
	if dataDirExistsAsFolder(parentOnly) {
		t.Fatal("expected false when only parent exists")
	}
}

func TestApp_requireReady_nilReceiver(t *testing.T) {
	var a *App
	if err := a.requireReady(); err == nil {
		t.Fatal("expected error")
	}
}
