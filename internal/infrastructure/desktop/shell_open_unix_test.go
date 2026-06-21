//go:build !windows && !darwin

package desktop

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenFileXDG_success(t *testing.T) {
	installFakeXDGOpen(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "doc.txt")
	if err := os.WriteFile(p, []byte("hi"), 0o600); err != nil {
		t.Fatal(err)
	}
	abs, err := ValidateRegularFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := openFileXDG(abs); err != nil {
		t.Fatalf("openFileXDG: %v", err)
	}
}

func TestOpenFileXDG_commandFails(t *testing.T) {
	installFailingXDGOpen(t)
	if err := openFileXDG("/tmp/file.txt"); err == nil {
		t.Fatal("expected error when xdg-open fails")
	}
}

func TestOpenFileXDG_notFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := openFileXDG("/tmp/file.txt"); err == nil {
		t.Fatal("expected error when xdg-open missing")
	}
}

func TestOpenFileWithDefaultApp_linuxXDG(t *testing.T) {
	installFakeXDGOpen(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := OpenFileWithDefaultApp(p); err != nil {
		t.Fatalf("OpenFileWithDefaultApp: %v", err)
	}
}

func TestOpenFolderInFileManager_linuxXDG(t *testing.T) {
	installFakeXDGOpen(t)
	dir := t.TempDir()
	if err := OpenFolderInFileManager(dir); err != nil {
		t.Fatalf("OpenFolderInFileManager: %v", err)
	}
}

func TestRevealInFileManager_linuxNautilus(t *testing.T) {
	installFakeNautilus(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "shot.png")
	if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := RevealInFileManager(p); err != nil {
		t.Fatalf("RevealInFileManager: %v", err)
	}
}

func TestRevealInFileManager_linuxXDGFallback(t *testing.T) {
	installFailingNautilus(t)
	installFakeXDGOpen(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "shot.png")
	if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := RevealInFileManager(p); err != nil {
		t.Fatalf("RevealInFileManager: %v", err)
	}
}

func TestRevealLinux_noFileManager(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	abs, err := ValidateRegularFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := revealLinux(abs); err == nil {
		t.Fatal("expected error when no file manager available")
	}
}

func TestRevealLinux_xdgOpenFails(t *testing.T) {
	installFailingNautilus(t)
	installFailingXDGOpen(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	abs, err := ValidateRegularFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := revealLinux(abs); err == nil {
		t.Fatal("expected error when xdg-open fails")
	}
}
