package vrchatwindow

import (
	"errors"
	"testing"
)

func TestResize_invalidSize(t *testing.T) {
	if err := Resize(0, 720); !errors.Is(err, ErrInvalidSize) {
		t.Fatalf("got %v, want ErrInvalidSize", err)
	}
	if err := Resize(1280, -1); !errors.Is(err, ErrInvalidSize) {
		t.Fatalf("got %v, want ErrInvalidSize", err)
	}
}

func TestResize_stubOrLive(t *testing.T) {
	err := Resize(1280, 720)
	// On Linux CI: unsupported. On Windows without VRChat: not running / no window.
	if err == nil {
		return
	}
	if errors.Is(err, ErrUnsupported) || errors.Is(err, ErrNotRunning) || errors.Is(err, ErrNoWindow) {
		return
	}
	t.Fatalf("unexpected error: %v", err)
}
