package vrchatapi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileCredentialStore_roundTrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "tok")
	s := NewFileCredentialStore(p)

	if _, err := s.Get(CredentialService, CredentialUser); err == nil {
		t.Fatal("Get: want error before Set")
	}

	if err := s.Set(CredentialService, CredentialUser, "secret-token"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := s.Get(CredentialService, CredentialUser)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "secret-token" {
		t.Fatalf("Get: want %q, got %q", "secret-token", got)
	}

	if err := s.Delete(CredentialService, CredentialUser); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := s.Get(CredentialService, CredentialUser); err == nil {
		t.Fatal("Get after Delete: want error")
	}

	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatalf("after Delete: want file absent, stat err=%v", err)
	}
}
