package logwatcher

import (
	"errors"
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

func TestListOutputLogFiles_ExcludesParsedLinesAuxiliary(t *testing.T) {
	dir := t.TempDir()
	realPath := filepath.Join(dir, "output_log_2026-03-18_12-52-26.txt")
	auxPath := filepath.Join(dir, "output_log_2026-03-18_12-52-26.parsed_lines.txt")
	if err := os.WriteFile(realPath, []byte("full"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(auxPath, []byte("subset"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := ListOutputLogFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != realPath {
		t.Fatalf("got %v, want single %q", got, realPath)
	}
}

func TestResolveLatestOutputLogFile_IgnoresParsedLinesAuxiliary(t *testing.T) {
	dir := t.TempDir()
	realPath := filepath.Join(dir, "output_log_2026-03-18_12-52-26.txt")
	auxPath := filepath.Join(dir, "output_log_2026-03-18_12-52-26.parsed_lines.txt")
	if err := os.WriteFile(realPath, []byte("full"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(auxPath, []byte("subset"), 0600); err != nil {
		t.Fatal(err)
	}
	auxT := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	realT := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	if err := os.Chtimes(auxPath, auxT, auxT); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(realPath, realT, realT); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveLatestOutputLogFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != realPath {
		t.Fatalf("got %q, want %q (must not pick newer .parsed_lines)", got, realPath)
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

func TestListOutputLogFiles_EmptyDirectory(t *testing.T) {
	_, err := ListOutputLogFiles(t.TempDir())
	if !errors.Is(err, ErrNoOutputLogFiles) {
		t.Fatalf("err = %v", err)
	}
}

func TestOutputLogPathValid_missingPath(t *testing.T) {
	if OutputLogPathValid(filepath.Join(t.TempDir(), "nope")) {
		t.Fatal("missing path should be invalid")
	}
}

func TestResolveLatestOutputLogFile_skipsDirectoryNamedLikeLog(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "output_log_2026-01-01_00-00-00.txt")
	if err := os.WriteFile(target, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "output_log_dir.txt"), 0755); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveLatestOutputLogFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != target {
		t.Fatalf("got %q want %q", got, target)
	}
}

func TestListOutputLogFiles_onlyAuxiliaryFiles(t *testing.T) {
	dir := t.TempDir()
	aux := filepath.Join(dir, "output_log_2026-03-18_12-52-26.parsed_lines.txt")
	if err := os.WriteFile(aux, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := ListOutputLogFiles(dir)
	if !errors.Is(err, ErrNoOutputLogFiles) {
		t.Fatalf("err = %v", err)
	}
}

func TestResolveLatestOutputLogFile_onlyBrokenSymlink(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "output_log_broken.txt")
	if err := os.Symlink(filepath.Join(dir, "missing-target.txt"), link); err != nil {
		t.Skip("symlink unsupported")
	}
	_, err := ResolveLatestOutputLogFile(dir)
	if !errors.Is(err, ErrNoOutputLogFiles) {
		t.Fatalf("err = %v", err)
	}
}
