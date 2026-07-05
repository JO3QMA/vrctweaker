package vrchatapi

import (
	"testing"
)

func TestStubCredentialStore_roundTrip(t *testing.T) {
	s := NewStubCredentialStore()

	if _, err := s.Get(CredentialService, CredentialUser); err == nil {
		t.Fatal("Get before Set: want error")
	}

	if err := s.Set(CredentialService, CredentialUser, "tok"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := s.Get(CredentialService, CredentialUser)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "tok" {
		t.Fatalf("Get = %q, want tok", got)
	}

	if err := s.Delete(CredentialService, CredentialUser); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := s.Get(CredentialService, CredentialUser); err == nil {
		t.Fatal("Get after Delete: want error")
	}
}
