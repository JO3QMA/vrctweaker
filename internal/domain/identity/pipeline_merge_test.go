package identity

import (
	"testing"
	"time"
)

func TestPipelineLocationIsHidden(t *testing.T) {
	t.Parallel()
	if !PipelineLocationIsHidden("private", "") {
		t.Fatal("worldID private should hide")
	}
	if !PipelineLocationIsHidden("", "private") {
		t.Fatal("location private should hide")
	}
	if PipelineLocationIsHidden("", "") {
		t.Fatal("empty should not hide via private check")
	}
	if PipelineLocationIsHidden("", "wrld_x:123") {
		t.Fatal("concrete location should not hide")
	}
}

func TestUserCache_MergeFromPipelineFriendOnline(t *testing.T) {
	t.Parallel()
	now := time.Unix(1700000000, 0)
	u := &UserCache{VRCUserID: "usr_1", UserKind: UserKindFriend, IsFavorite: true}
	u.MergeFromPipelineFriendOnline(now, "standalonewindows", "wrld_a:inst~id", false)
	if u.Location != "wrld_a:inst~id" {
		t.Fatalf("location: got %q", u.Location)
	}
	if u.Platform != "standalonewindows" {
		t.Fatalf("platform: got %q", u.Platform)
	}
	if u.LastUpdated != now {
		t.Fatal("LastUpdated")
	}

	u2 := &UserCache{VRCUserID: "usr_2", UserKind: UserKindFriend}
	u2.MergeFromPipelineFriendOnline(now, "standalonewindows", "", false)
	if u2.Location != PipelineLocationUnknown {
		t.Fatalf("empty location: got %q", u2.Location)
	}

	u3 := &UserCache{VRCUserID: "usr_3", UserKind: UserKindFriend}
	u3.MergeFromPipelineFriendOnline(now, "android", "private", false)
	if u3.Location != PipelineLocationUnknown {
		t.Fatalf("private location: got %q", u3.Location)
	}
}

func TestUserCache_MergeFromPipelineFriendLocation(t *testing.T) {
	t.Parallel()
	now := time.Unix(1700000001, 0)
	u := &UserCache{VRCUserID: "usr_1", UserKind: UserKindFriend}
	u.MergeFromPipelineFriendLocation(now, "traveling", "wrld_dest:abc", "")
	if u.Location != "wrld_dest:abc" {
		t.Fatalf("traveling dest: got %q", u.Location)
	}

	u2 := &UserCache{VRCUserID: "usr_2", UserKind: UserKindFriend}
	u2.MergeFromPipelineFriendLocation(now, "wrld_x:1", "", "private")
	if u2.Location != PipelineLocationUnknown {
		t.Fatalf("private worldId: got %q", u2.Location)
	}
}

func TestUserCache_MergeFromPipelineFriendOffline(t *testing.T) {
	t.Parallel()
	now := time.Unix(1700000002, 0)
	u := &UserCache{
		VRCUserID: "usr_1", UserKind: UserKindFriend,
		Status: "active", Platform: "web", Location: "wrld_x:1",
	}
	u.MergeFromPipelineFriendOffline(now)
	if u.Status != "offline" || u.Platform != "" || u.Location != "" {
		t.Fatalf("got status=%q platform=%q loc=%q", u.Status, u.Platform, u.Location)
	}
}

func TestUserCache_MergeFromPipelineFriendUser_preservesFavorite(t *testing.T) {
	t.Parallel()
	now := time.Unix(1700000003, 0)
	u := &UserCache{VRCUserID: "usr_1", UserKind: UserKindFriend, IsFavorite: true}
	snap := &UserCache{
		VRCUserID: "usr_1", DisplayName: "N", Status: "active", UserKind: UserKindFriend,
		IsFavorite: false, LastUpdated: now,
	}
	u.MergeFromPipelineFriendUser(snap, now)
	if !u.IsFavorite {
		t.Fatal("IsFavorite should be preserved")
	}
	if u.DisplayName != "N" || u.Status != "active" {
		t.Fatalf("profile not merged: %+v", u)
	}
}

func TestUserCache_DemoteFriendToContactAfterUnfriend(t *testing.T) {
	t.Parallel()
	now := time.Unix(1700000004, 0)
	u := &UserCache{
		VRCUserID: "usr_1", UserKind: UserKindFriend, IsFavorite: true,
		Status: "active", Location: "wrld_x:1",
	}
	u.DemoteFriendToContactAfterUnfriend(now)
	if u.UserKind != UserKindContact || u.IsFavorite {
		t.Fatalf("got kind=%v fav=%v", u.UserKind, u.IsFavorite)
	}
	if u.Status != "" || u.Location != "" {
		t.Fatalf("presence should clear: status=%q loc=%q", u.Status, u.Location)
	}
}

func TestUserCache_MergeFromPipelineSelfUserUpdate_ignoresNonSelf(t *testing.T) {
	t.Parallel()
	now := time.Unix(1700000005, 0)
	u := &UserCache{VRCUserID: "usr_1", UserKind: UserKindFriend, DisplayName: "F"}
	u.MergeFromPipelineSelfUserUpdate(now, "X", "join me", "", "", "", "", "")
	if u.DisplayName != "F" {
		t.Fatal("friend row should not update")
	}
}

func TestUserCache_MergeFromPipelineSelfLocation(t *testing.T) {
	t.Parallel()
	now := time.Unix(1700000006, 0)
	u := &UserCache{VRCUserID: "usr_me", UserKind: UserKindSelf, DisplayName: "Me"}
	u.MergeFromPipelineSelfLocation(now, "wrld_a:inst", "")
	if u.Location != "wrld_a:inst" {
		t.Fatalf("location: %q", u.Location)
	}
	u.MergeFromPipelineSelfLocation(now, "private", "")
	if u.Location != PipelineLocationUnknown {
		t.Fatalf("private: %q", u.Location)
	}
}
