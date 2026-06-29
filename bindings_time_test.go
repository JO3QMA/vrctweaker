package main

import (
	"testing"
	"time"
)

func TestParseDateOrRFC3339(t *testing.T) {
	t.Parallel()
	day := parseDateOrRFC3339("2024-06-01")
	if day == nil {
		t.Fatal("expected date")
	}
	if day.Format("2006-01-02") != "2024-06-01" {
		t.Fatalf("got %s", day.Format(time.RFC3339))
	}
	rfc := parseDateOrRFC3339("2024-06-01T12:00:00Z")
	if rfc == nil || !rfc.Equal(time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("got %v", rfc)
	}
	if parseDateOrRFC3339("") != nil {
		t.Fatal("empty should be nil")
	}
}
