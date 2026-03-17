//go:build integration

package vrchatapi

import (
	"testing"
)

func TestKeyringCredentialStore_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping keyring integration test in short mode")
	}
	store := NewKeyringCredentialStore()
	service := CredentialService + "-test"
	user := CredentialUser + "-test"
	token := "test-auth-token-" + "12345"

	// Clean up in case of previous run
	_ = store.Delete(service, user)

	if err := store.Set(service, user, token); err != nil {
		t.Fatalf("Set: %v (keyring may be unavailable in CI)", err)
	}
	defer func() { _ = store.Delete(service, user) }()

	got, err := store.Get(service, user)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != token {
		t.Errorf("Get: want %q, got %q", token, got)
	}

	if err := store.Delete(service, user); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err = store.Get(service, user)
	if err == nil {
		t.Error("Get after Delete: want error, got nil")
	}
}
