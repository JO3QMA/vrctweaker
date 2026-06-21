package identity

import (
	"testing"
	"time"
)

func TestUserCache_MergeFromPipelineFriendOnline_hideLocation(t *testing.T) {
	t.Parallel()
	now := time.Unix(1700000100, 0)
	u := &UserCache{VRCUserID: "usr_h", UserKind: UserKindFriend}
	u.MergeFromPipelineFriendOnline(now, "standalonewindows", "wrld_visible:1", true)
	if u.Location != PipelineLocationUnknown {
		t.Fatalf("hideLocation: got %q", u.Location)
	}
}

func TestUserCache_MergeFromPipelineFriendOnline_nilSelfIgnored(t *testing.T) {
	t.Parallel()
	var u *UserCache
	u.MergeFromPipelineFriendOnline(time.Now(), "web", "loc", false)
}

func TestUserCache_MergeFromPipelineFriendLocation_emptyLocation(t *testing.T) {
	t.Parallel()
	now := time.Unix(1700000101, 0)
	u := &UserCache{VRCUserID: "usr_e", UserKind: UserKindFriend}
	u.MergeFromPipelineFriendLocation(now, "", "", "")
	if u.Location != PipelineLocationUnknown {
		t.Fatalf("got %q", u.Location)
	}
}

func TestUserCache_MergeFromPipelineFriendLocation_travelingWithoutDest(t *testing.T) {
	t.Parallel()
	now := time.Unix(1700000102, 0)
	u := &UserCache{VRCUserID: "usr_t", UserKind: UserKindFriend}
	u.MergeFromPipelineFriendLocation(now, "traveling", "", "")
	if u.Location != "traveling" {
		t.Fatalf("got %q", u.Location)
	}
}

func TestUserCache_MergeFromPipelineFriendUser_nilSnap(t *testing.T) {
	t.Parallel()
	u := &UserCache{VRCUserID: "usr_1", UserKind: UserKindFriend, DisplayName: "Keep"}
	u.MergeFromPipelineFriendUser(nil, time.Now())
	if u.DisplayName != "Keep" {
		t.Fatal("should not change on nil snap")
	}
}

func TestUserCache_MergeFromPipelineSelfUserUpdate_allFields(t *testing.T) {
	t.Parallel()
	now := time.Unix(1700000103, 0)
	u := &UserCache{VRCUserID: "usr_me", UserKind: UserKindSelf}
	u.MergeFromPipelineSelfUserUpdate(now,
		"NewName", "join me", "desc", "user1",
		"avatar.png", "icon.png", "profile.png",
	)
	if u.DisplayName != "NewName" || u.Status != "join me" || u.StatusDescription != "desc" ||
		u.Username != "user1" || u.AvatarThumbnailURL != "avatar.png" ||
		u.UserIconURL != "icon.png" || u.ProfilePicOverrideThumbnail != "profile.png" {
		t.Fatalf("got %+v", u)
	}
	if u.LastUpdated != now {
		t.Fatal("LastUpdated")
	}
}

func TestUserCache_MergeFromPipelineSelfLocation_traveling(t *testing.T) {
	t.Parallel()
	now := time.Unix(1700000104, 0)
	u := &UserCache{VRCUserID: "usr_me", UserKind: UserKindSelf}
	u.MergeFromPipelineSelfLocation(now, "traveling", "wrld_dest:abc")
	if u.Location != "wrld_dest:abc" {
		t.Fatalf("got %q", u.Location)
	}
}

func TestUserCache_MergeFromPipelineSelfLocation_emptyKeepsPrevious(t *testing.T) {
	t.Parallel()
	now := time.Unix(1700000105, 0)
	u := &UserCache{VRCUserID: "usr_me", UserKind: UserKindSelf, Location: "wrld_keep:1"}
	u.MergeFromPipelineSelfLocation(now, "", "")
	if u.Location != "wrld_keep:1" {
		t.Fatalf("got %q", u.Location)
	}
}

func TestUserCache_MergeFromPipelineFriendOffline_ignoresSelf(t *testing.T) {
	t.Parallel()
	u := &UserCache{VRCUserID: "usr_me", UserKind: UserKindSelf, Status: "active"}
	u.MergeFromPipelineFriendOffline(time.Now())
	if u.Status != "active" {
		t.Fatal("self row should not change")
	}
}

func TestUserCache_DemoteFriendToContactAfterUnfriend_ignoresSelf(t *testing.T) {
	t.Parallel()
	u := &UserCache{VRCUserID: "usr_me", UserKind: UserKindSelf}
	u.DemoteFriendToContactAfterUnfriend(time.Now())
	if u.UserKind != UserKindSelf {
		t.Fatal("self row should not demote")
	}
}
