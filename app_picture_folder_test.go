package main

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestApp_DefaultVRChatPictureFolder(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", home)
	}

	a := NewApp()
	got, err := a.DefaultVRChatPictureFolder()
	if err != nil {
		t.Fatalf("DefaultVRChatPictureFolder: %v", err)
	}
	want := filepath.Join(home, "Pictures", "VRChat")
	if got != want {
		t.Errorf("DefaultVRChatPictureFolder() = %q, want %q", got, want)
	}
}
