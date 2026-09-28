package wailsapp

import (
	"testing"
	"time"
)

func TestFormatRFC3339_zeroTimeReturnsEmpty(t *testing.T) {
	if formatRFC3339(time.Time{}) != "" {
		t.Fatalf("zero time should map to empty string")
	}
}

func TestFormatRFC3339_nonZero(t *testing.T) {
	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	got := formatRFC3339(ts)
	if got != "2026-01-02T03:04:05Z" {
		t.Fatalf("got %q", got)
	}
}
