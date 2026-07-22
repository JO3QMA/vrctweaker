package vrchatwindow

import (
	"errors"
	"math"
	"testing"
)

func TestResize_invalidSize(t *testing.T) {
	cases := []struct {
		w, h int
	}{
		{0, 720},
		{1280, -1},
		{math.MaxInt32 + 1, 720},
		{1280, math.MaxInt32 + 1},
	}
	for _, tc := range cases {
		if err := Resize(tc.w, tc.h); !errors.Is(err, ErrInvalidSize) {
			t.Fatalf("Resize(%d,%d)=%v, want ErrInvalidSize", tc.w, tc.h, err)
		}
	}
}

func TestResize_stubOrLive(t *testing.T) {
	err := Resize(1280, 720)
	if err == nil {
		return
	}
	if errors.Is(err, ErrUnsupported) ||
		errors.Is(err, ErrNotRunning) ||
		errors.Is(err, ErrNoWindow) ||
		errors.Is(err, ErrMultipleInstances) {
		return
	}
	t.Fatalf("unexpected error: %v", err)
}
