package desktop

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateRegularFile_OK(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := ValidateRegularFile(p)
	if err != nil {
		t.Fatalf("ValidateRegularFile: %v", err)
	}
	if got == "" {
		t.Fatal("empty abs path")
	}
	info, err := os.Stat(got)
	if err != nil || !info.Mode().IsRegular() {
		t.Fatalf("expected regular file at %q", got)
	}
}

func TestValidateRegularFile_Empty(t *testing.T) {
	t.Parallel()
	_, err := ValidateRegularFile("")
	if err == nil {
		t.Fatal("expected error")
	}
	_, err = ValidateRegularFile("   ")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateRegularFile_NotFound(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, err := ValidateRegularFile(filepath.Join(dir, "nope.txt"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateRegularFile_NotRegular(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := ValidateRegularFile(filepath.Join(dir, "sub"))
	if err == nil {
		t.Fatal("expected error for directory")
	}
}
