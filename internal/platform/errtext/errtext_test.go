package errtext

import "testing"

func TestTruncateForErr_respectsRuneBoundary(t *testing.T) {
	t.Parallel()
	if got := TruncateForErr("日本語テスト", 3); got != "日本語…" {
		t.Fatalf("got %q", got)
	}
}

func TestTruncateForErr_shortAndTrim(t *testing.T) {
	t.Parallel()
	if got := TruncateForErr("  ok  ", 10); got != "ok" {
		t.Fatalf("got %q", got)
	}
	if got := TruncateForErr("short", 10); got != "short" {
		t.Fatalf("got %q", got)
	}
}

func TestTruncateForErr_nonPositiveMax(t *testing.T) {
	t.Parallel()
	if got := TruncateForErr("  ok  ", 0); got != "ok" {
		t.Fatalf("got %q", got)
	}
	if got := TruncateForErr("ok", -1); got != "ok" {
		t.Fatalf("got %q", got)
	}
}
