package logwatcher

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestListOutputLogFiles_EmptyDirectory(t *testing.T) {
	_, err := ListOutputLogFiles(t.TempDir())
	if !errors.Is(err, ErrNoOutputLogFiles) {
		t.Fatalf("err = %v", err)
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

func TestOutputLogPathValid_emptyDirOK(t *testing.T) {
	dir := t.TempDir()
	if !OutputLogPathValid(dir) {
		t.Fatal("empty dir should be valid")
	}
}

func TestOutputLogPathValid_dirWithLogs(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "output_log_1.txt")
	if err := os.WriteFile(p, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if !OutputLogPathValid(dir) {
		t.Fatal("dir with logs should be valid")
	}
}

func TestOutputLogPathValid_fileRejected(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "output_log_1.txt")
	if err := os.WriteFile(p, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if OutputLogPathValid(p) {
		t.Fatal("regular file should be rejected")
	}
}

func TestOutputLogPathValid_missingPath(t *testing.T) {
	if OutputLogPathValid(filepath.Join(t.TempDir(), "nope")) {
		t.Fatal("missing path should be invalid")
	}
}
