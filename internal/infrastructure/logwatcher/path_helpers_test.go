package logwatcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBootstrapLiveLogFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "output_log_2026-03-20_23-00-00.txt")
	midPath := filepath.Join(dir, "output_log_2026-03-21_08-00-00.txt")
	newPath := filepath.Join(dir, "output_log_2026-03-21_11-00-00.txt")

	base := time.Date(2026, 3, 21, 11, 0, 0, 0, time.UTC)
	for _, spec := range []struct {
		path string
		mod  time.Time
	}{
		{oldPath, base.Add(-12 * time.Hour)},
		{midPath, base.Add(-3 * time.Hour)},
		{newPath, base},
	} {
		if err := os.WriteFile(spec.path, []byte("log\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(spec.path, spec.mod, spec.mod); err != nil {
			t.Fatal(err)
		}
	}

	live := BootstrapLiveLogFiles([]string{oldPath, midPath, newPath}, 5*time.Second)
	if live == nil {
		t.Fatal("expected live map for directory with recent tail file")
	}
	if live[oldPath] || live[midPath] {
		t.Fatalf("unexpected live flags: old=%v mid=%v", live[oldPath], live[midPath])
	}
	if !live[newPath] {
		t.Fatal("expected newest file to be live")
	}
}

func TestMatchAbsPaths(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "./a.txt")
	if !MatchAbsPaths(a, b) {
		t.Fatalf("MatchAbsPaths(%q, %q) = false, want true", a, b)
	}
}
