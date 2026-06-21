package vrchatapi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewAutoCredentialStore_forcesFileWhenEnvSet(t *testing.T) {
	t.Setenv("VRCHAT_TWEAKER_USE_FILE_CREDENTIALS", "1")
	dir := t.TempDir()
	s := NewAutoCredentialStore(dir, nil)

	if err := s.Set(CredentialService, CredentialUser, "file-tok"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := s.Get(CredentialService, CredentialUser)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "file-tok" {
		t.Fatalf("Get = %q, want file-tok", got)
	}

	p := filepath.Join(dir, ".vrchat-auth-token")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("expected file store at %s: %v", p, err)
	}
}

func TestNewAutoCredentialStore_returnsStore(t *testing.T) {
	t.Setenv("VRCHAT_TWEAKER_USE_FILE_CREDENTIALS", "")
	dir := t.TempDir()
	var warned bool
	s := NewAutoCredentialStore(dir, func(string) { warned = true })
	if s == nil {
		t.Fatal("NewAutoCredentialStore returned nil")
	}
	if err := s.Set(CredentialService, CredentialUser, "x"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	_ = warned // keyring or file path; either is fine
}
