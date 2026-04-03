package vrchatapi

import (
	"errors"
	"path/filepath"
	"testing"
)

type stubFailingSet struct{}

func (stubFailingSet) Get(string, string) (string, error) { return "", errors.New("not found") }

func (stubFailingSet) Set(string, string, string) error {
	return errors.New("exec: \"dbus-launch\": executable file not found in $PATH")
}

func (stubFailingSet) Delete(string, string) error { return nil }

func TestCredentialStoreWithFileFallback_SetUsesFileWhenPrimaryFails(t *testing.T) {
	dir := t.TempDir()
	fileStore := NewFileCredentialStore(filepath.Join(dir, "tok"))
	var warned string
	w := func(msg string) { warned = msg }
	s := newCredentialStoreWithFileFallback(stubFailingSet{}, fileStore, w)

	if err := s.Set(CredentialService, CredentialUser, "tok123"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if warned == "" {
		t.Fatal("expected warn callback")
	}

	got, err := s.Get(CredentialService, CredentialUser)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "tok123" {
		t.Fatalf("Get: want tok123, got %q", got)
	}
}
