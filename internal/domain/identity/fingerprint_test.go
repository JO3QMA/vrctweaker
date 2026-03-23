package identity

import "testing"

func TestAuthTokenFingerprint_stable(t *testing.T) {
	a := AuthTokenFingerprint("same-token")
	b := AuthTokenFingerprint("same-token")
	if a == "" || a != b {
		t.Fatalf("fingerprint = %q %q", a, b)
	}
	if AuthTokenFingerprint("other") == a {
		t.Fatal("different tokens should not match")
	}
	if AuthTokenFingerprint("") != "" {
		t.Fatal("empty token want empty fingerprint")
	}
}
