package wailsapp

import (
	"testing"
)

func TestOpenExternalURL_uninitialized(t *testing.T) {
	var a App
	if err := a.OpenExternalURL("https://example.com"); err == nil {
		t.Fatal("expected error when ctx is nil")
	}
}

func TestOpenExternalURL_rejectsUnsafe(t *testing.T) {
	a := &App{}
	if err := a.OpenExternalURL("javascript:alert(1)"); err == nil {
		t.Fatal("expected rejection")
	}
}
