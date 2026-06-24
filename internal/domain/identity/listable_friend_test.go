package identity

import "testing"

func TestIsListableFriend(t *testing.T) {
	t.Parallel()
	if IsListableFriend(nil) {
		t.Fatal("nil is not listable")
	}
	if IsListableFriend(&UserCache{UserKind: UserKindFriend, DisplayName: "A"}) != true {
		t.Fatal("named friend should be listable")
	}
	if IsListableFriend(&UserCache{UserKind: UserKindFriend, DisplayName: "  "}) {
		t.Fatal("blank display name should not be listable")
	}
	if IsListableFriend(&UserCache{UserKind: UserKindContact, DisplayName: "A"}) {
		t.Fatal("contact should not be listable")
	}
}
