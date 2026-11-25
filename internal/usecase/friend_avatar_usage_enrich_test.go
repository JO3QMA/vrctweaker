package usecase

import (
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/identity"
)

func TestEnrichFriendAvatarUsages_matchesCurrentAvatarName(t *testing.T) {
	now := time.Now()
	list := []*activity.FriendAvatarUsageSummary{
		{AvatarName: "Legacy Fit", UseCount: 2, FirstSeenAt: now, LastSeenAt: now},
		{AvatarName: "Current One", UseCount: 1, FirstSeenAt: now, LastSeenAt: now},
	}
	friend := &identity.UserCache{CurrentAvatarID: "avtr_11111111-2222-3333-4444-555555555555"}
	out := EnrichFriendAvatarUsages(list, friend, "Current One")
	if out[0].AvatarID != "" {
		t.Fatalf("legacy row id: %q", out[0].AvatarID)
	}
	if out[1].AvatarID != friend.CurrentAvatarID {
		t.Fatalf("current row id: %q", out[1].AvatarID)
	}
}

func TestEnrichFriendAvatarUsages_skipsNilRows(t *testing.T) {
	now := time.Now()
	list := []*activity.FriendAvatarUsageSummary{
		nil,
		{AvatarName: "Keep", UseCount: 1, FirstSeenAt: now, LastSeenAt: now},
		nil,
	}
	out := EnrichFriendAvatarUsages(list, nil, "")
	if len(out) != 1 {
		t.Fatalf("len %d want 1", len(out))
	}
	if out[0].AvatarName != "Keep" {
		t.Fatalf("name %q", out[0].AvatarName)
	}
}
