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

// variablePrimary is a test double with togglable Set/Get failure and an in-memory secret on successful Set.
type variablePrimary struct {
	failSet bool
	failGet bool
	secret  string
}

func (v *variablePrimary) Get(string, string) (string, error) {
	if v.failGet {
		return "", errors.New("dbus get failed")
	}
	if v.secret == "" {
		return "", errors.New("not found")
	}
	return v.secret, nil
}

func (v *variablePrimary) Set(_, _, password string) error {
	if v.failSet {
		return errors.New("dbus set failed")
	}
	v.secret = password
	return nil
}

func (*variablePrimary) Delete(string, string) error { return nil }

// Regression: after keyring Set succeeds, stale file fallback from an earlier failed Set must not be used when Get falls back to file.
func TestCredentialStoreWithFileFallback_SetClearsFileAfterPrimarySucceeds(t *testing.T) {
	dir := t.TempDir()
	fileStore := NewFileCredentialStore(filepath.Join(dir, "tok"))
	p := &variablePrimary{failSet: true, failGet: true}
	s := newCredentialStoreWithFileFallback(p, fileStore, func(string) {})

	if err := s.Set(CredentialService, CredentialUser, "tokenA"); err != nil {
		t.Fatalf("first Set (file fallback): %v", err)
	}

	p.failSet = false
	if err := s.Set(CredentialService, CredentialUser, "tokenB"); err != nil {
		t.Fatalf("second Set (keyring): %v", err)
	}

	_, err := s.Get(CredentialService, CredentialUser)
	if err == nil {
		t.Fatal("Get: expected error when keyring Get fails and file fallback was cleared, got nil")
	}
}

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
