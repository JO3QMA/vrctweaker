package vrchatapi

import "testing"

func TestNewKeyringCredentialStore(t *testing.T) {
	if NewKeyringCredentialStore() == nil {
		t.Fatal("NewKeyringCredentialStore returned nil")
	}
}

func TestKeyringCredentialStore_GetAfterSet(t *testing.T) {
	store := NewKeyringCredentialStore()
	service := CredentialService + "-unit-get"
	user := CredentialUser + "-unit-get"
	_ = store.Delete(service, user)

	if err := store.Set(service, user, "tok"); err != nil {
		t.Skipf("keyring Set unavailable: %v", err)
	}
	t.Cleanup(func() { _ = store.Delete(service, user) })

	got, err := store.Get(service, user)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "tok" {
		t.Fatalf("Get = %q, want tok", got)
	}
}
