package wailsapp

import "testing"

func TestMinWindowBounds(t *testing.T) {
	if MinWindowWidth != 768 {
		t.Fatalf("MinWindowWidth = %d, want 768", MinWindowWidth)
	}
	if MinWindowHeight != 540 {
		t.Fatalf("MinWindowHeight = %d, want 540", MinWindowHeight)
	}
}
