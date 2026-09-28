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
	out := EnrichFriendAvatarUsages(list, friend, "Current One", "/cache")
	if out[0].AvatarID != "" {
		t.Fatalf("legacy row id: %q", out[0].AvatarID)
	}
	if out[1].AvatarID != friend.CurrentAvatarID {
		t.Fatalf("current row id: %q", out[1].AvatarID)
	}
	if out[1].LocalCachePath != "" {
		t.Fatalf("path not resolved in enrich without finder: %q", out[1].LocalCachePath)
	}
}

func TestEnrichFriendAvatarUsages_resolvesCachePath(t *testing.T) {
	now := time.Now()
	list := []*activity.FriendAvatarUsageSummary{
		{AvatarName: "A", UseCount: 1, FirstSeenAt: now, LastSeenAt: now},
	}
	id := "avtr_aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	friend := &identity.UserCache{CurrentAvatarID: id}
	out := EnrichFriendAvatarUsages(list, friend, "A", "/cache", func(root, avatarID string) (string, error) {
		if root == "/cache" && avatarID == id {
			return "/cache/found", nil
		}
		return "", nil
	})
	if out[0].LocalCachePath != "/cache/found" {
		t.Fatalf("path %q", out[0].LocalCachePath)
	}
}
