package logwatcher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestListOutputLogFiles_Sorted(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "output_log_2026-01-01_00-00-00.txt"), []byte("a"), 0600)
	_ = os.WriteFile(filepath.Join(dir, "output_log_2026-03-22_00-00-00.txt"), []byte("b"), 0600)
	got, err := ListOutputLogFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d", len(got))
	}
	if !strings.Contains(filepath.Base(got[0]), "2026-01-01") {
		t.Errorf("first = %s", got[0])
	}
}

func TestResolveLatestOutputLogFile_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	_, err := ResolveLatestOutputLogFile(dir)
	if err == nil {
		t.Fatal("expected error")
	}
	if err != ErrNoOutputLogFiles {
		t.Fatalf("err = %v, want ErrNoOutputLogFiles", err)
	}
}

func TestResolveLatestOutputLogFile_IgnoresNonMatching(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "other.txt"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := ResolveLatestOutputLogFile(dir)
	if err != ErrNoOutputLogFiles {
		t.Fatalf("err = %v, want ErrNoOutputLogFiles", err)
	}
}

func TestResolveLatestOutputLogFile_PicksNewestModTime(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "output_log_2026-01-01_00-00-00.txt")
	newPath := filepath.Join(dir, "output_log_2026-03-22_00-47-45.txt")
	if err := os.WriteFile(oldPath, []byte("a"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte("b"), 0600); err != nil {
		t.Fatal(err)
	}
	oldT := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newT := time.Date(2026, 3, 22, 0, 47, 45, 0, time.UTC)
	if err := os.Chtimes(oldPath, oldT, oldT); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(newPath, newT, newT); err != nil {
		t.Fatal(err)
	}

	got, err := ResolveLatestOutputLogFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != newPath {
		t.Fatalf("got %q, want %q", got, newPath)
	}
}

func TestOutputLogPathValid(t *testing.T) {
	dir := t.TempDir()
	if OutputLogPathValid(dir) {
		t.Error("empty dir should be invalid")
	}
	p := filepath.Join(dir, "output_log_1.txt")
	if err := os.WriteFile(p, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if !OutputLogPathValid(dir) {
		t.Error("dir with output_log*.txt should be valid")
	}
	if !OutputLogPathValid(p) {
		t.Error("regular log file should be valid")
	}
	other := filepath.Join(dir, "readme.txt")
	if err := os.WriteFile(other, []byte("y"), 0600); err != nil {
		t.Fatal(err)
	}
	if !OutputLogPathValid(other) {
		t.Error("any existing regular file is accepted as explicit log path")
	}
}

func TestResolveLatestOutputLogFile_SameModTime_NameTieBreak(t *testing.T) {
	dir := t.TempDir()
	aPath := filepath.Join(dir, "output_log_a.txt")
	bPath := filepath.Join(dir, "output_log_b.txt")
	if err := os.WriteFile(aPath, []byte("1"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bPath, []byte("2"), 0600); err != nil {
		t.Fatal(err)
	}
	ts := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	if err := os.Chtimes(aPath, ts, ts); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(bPath, ts, ts); err != nil {
		t.Fatal(err)
	}

	got, err := ResolveLatestOutputLogFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != bPath {
		t.Fatalf("same mtime: got %q, want %q (lexicographic last)", got, bPath)
	}
}
