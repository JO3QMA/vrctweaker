package wailsapp

import (
	"context"
	"testing"
)

func TestApp_FriendAvatarUsageByVRCUserID_nilIdentityDoesNotPanic(t *testing.T) {
	app, _ := newTestAppWithActivity(t)
	app.ctx = context.Background()
	app.identity = nil

	got, err := app.FriendAvatarUsageByVRCUserID("usr_test")
	if err != nil {
		t.Fatalf("FriendAvatarUsageByVRCUserID: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil slice")
	}
	if len(got) != 0 {
		t.Fatalf("len %d want 0", len(got))
	}
}
