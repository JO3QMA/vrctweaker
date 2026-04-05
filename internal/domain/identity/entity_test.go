package identity

import (
	"testing"
	"time"
)

func TestUserCache_MergeFromAPIFriend_contactBecomesFriend(t *testing.T) {
	at := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	u := &UserCache{
		VRCUserID:   "usr_1",
		DisplayName: "Old",
		UserKind:    UserKindContact,
	}
	api := &UserCache{
		VRCUserID:   "usr_1",
		DisplayName: "FromAPI",
		Status:      "active",
		IsFavorite:  true,
		LastUpdated: at,
		UserKind:    UserKindFriend,
	}
	u.MergeFromAPIFriend(api)
	if u.UserKind != UserKindFriend {
		t.Fatalf("UserKind = %q, want friend", u.UserKind)
	}
	if u.DisplayName != "FromAPI" || u.Status != "active" || !u.IsFavorite {
		t.Fatalf("merged fields: %+v", u)
	}
	if !u.LastUpdated.Equal(at) {
		t.Fatalf("LastUpdated = %v, want %v", u.LastUpdated, at)
	}
}

func TestUserCache_MergeFromLog_doesNotDemoteFriendOrSelf(t *testing.T) {
	t1 := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	t2 := t1.Add(time.Hour)
	lu := time.Date(2026, 3, 19, 10, 0, 0, 0, time.UTC)

	friend := &UserCache{
		VRCUserID:     "f1",
		DisplayName:   "F",
		UserKind:      UserKindFriend,
		Status:        "offline",
		LastUpdated:   lu,
		FirstSeenAt:   ptrTime(t1),
		LastContactAt: ptrTime(t1),
	}
	friend.MergeFromLog("FLog", t2)
	if friend.UserKind != UserKindFriend {
		t.Fatalf("friend demoted to %q", friend.UserKind)
	}
	if !friend.LastUpdated.Equal(lu) {
		t.Fatalf("friend LastUpdated changed: %v want %v", friend.LastUpdated, lu)
	}
	if friend.FirstSeenAt == nil || !friend.FirstSeenAt.Equal(t1) {
		t.Fatalf("friend FirstSeenAt should stay first set: %v", friend.FirstSeenAt)
	}

	self := &UserCache{
		VRCUserID:     "me",
		DisplayName:   "Me",
		UserKind:      UserKindSelf,
		LastUpdated:   lu,
		FirstSeenAt:   ptrTime(t1),
		LastContactAt: ptrTime(t1),
	}
	self.MergeFromLog("MeLog", t2)
	if self.UserKind != UserKindSelf {
		t.Fatalf("self demoted to %q", self.UserKind)
	}
	if !self.LastUpdated.Equal(lu) {
		t.Fatalf("self LastUpdated changed: %v", self.LastUpdated)
	}
}

func TestUserCache_MergeFromAPIFriend_copiesExtendedFriendFields(t *testing.T) {
	at := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	u := &UserCache{
		VRCUserID:   "usr_1",
		UserKind:    UserKindContact,
		IsFavorite:  true,
		DisplayName: "Old",
	}
	api := &UserCache{
		VRCUserID:                   "usr_1",
		DisplayName:                 "API",
		Status:                      "active",
		IsFavorite:                  true,
		LastUpdated:                 at,
		UserKind:                    UserKindFriend,
		Username:                    "u1",
		StatusDescription:           "sd",
		UserState:                   "st",
		AvatarThumbnailURL:          "https://thumb",
		UserIconURL:                 "https://icon",
		ProfilePicOverrideThumbnail: "https://ppo/t",
		Bio:                         "bio text",
		BioLinksJSON:                `["https://a"]`,
		CurrentAvatarImageURL:       "https://full",
		CurrentAvatarTagsJSON:       `["t1"]`,
		DeveloperType:               "none",
		FriendKey:                   "fk",
		ImageURL:                    "https://img",
		LastPlatform:                "win",
		Location:                    "wrld:x",
		LastLogin:                   "2020-01-01T00:00:00Z",
		LastActivity:                "2020-01-02T00:00:00Z",
		LastMobile:                  "2020-01-03T00:00:00Z",
		Platform:                    "standalonewindows",
		ProfilePicOverride:          "https://ppo",
		TagsJSON:                    `["tag_a"]`,
	}
	u.MergeFromAPIFriend(api)
	if u.Bio != "bio text" || u.Location != "wrld:x" || u.BioLinksJSON != `["https://a"]` {
		t.Fatalf("extended fields: %+v", u)
	}
	if u.AvatarThumbnailURL != "https://thumb" || u.TagsJSON != `["tag_a"]` {
		t.Fatalf("thumb/tags: %+v", u)
	}
}

func TestUserCache_MergeFromAPIFriend_doesNotOverwriteSelf(t *testing.T) {
	at := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	u := &UserCache{
		VRCUserID:   "me",
		DisplayName: "Original",
		Status:      "busy",
		UserKind:    UserKindSelf,
		LastUpdated: at.Add(-time.Hour),
	}
	api := &UserCache{
		VRCUserID:   "me",
		DisplayName: "Hacked",
		Status:      "join me",
		LastUpdated: at,
	}
	u.MergeFromAPIFriend(api)
	if u.DisplayName != "Original" || u.Status != "busy" {
		t.Fatalf("self row overwritten: %+v", u)
	}
}

func TestUserCache_MergeFromLog_contactFirstSeenAndLastContact(t *testing.T) {
	tOld := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	tMid := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	tNew := time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

	u := &UserCache{VRCUserID: "c1", UserKind: UserKindContact}
	u.MergeFromLog("A", tMid)
	if u.FirstSeenAt == nil || !u.FirstSeenAt.Equal(tMid) {
		t.Fatalf("FirstSeenAt = %v, want %v", u.FirstSeenAt, tMid)
	}
	if u.LastContactAt == nil || !u.LastContactAt.Equal(tMid) {
		t.Fatalf("LastContactAt = %v, want %v", u.LastContactAt, tMid)
	}

	u.MergeFromLog("A", tNew)
	if u.FirstSeenAt == nil || !u.FirstSeenAt.Equal(tMid) {
		t.Fatalf("FirstSeenAt should stay earliest %v, got %v", tMid, u.FirstSeenAt)
	}
	if u.LastContactAt == nil || !u.LastContactAt.Equal(tNew) {
		t.Fatalf("LastContactAt = %v, want %v", u.LastContactAt, tNew)
	}

	u.MergeFromLog("A", tOld)
	if u.FirstSeenAt == nil || !u.FirstSeenAt.Equal(tOld) {
		t.Fatalf("FirstSeenAt = %v, want earlier %v", u.FirstSeenAt, tOld)
	}
	if u.LastContactAt == nil || !u.LastContactAt.Equal(tNew) {
		t.Fatalf("LastContactAt regressed: %v, want %v", u.LastContactAt, tNew)
	}
}

func TestUserCache_MergeFromLog_newRowSetsContactTimestamps(t *testing.T) {
	at := time.Date(2026, 3, 21, 15, 0, 0, 0, time.UTC)
	u := &UserCache{VRCUserID: "new"}
	u.MergeFromLog("NewName", at)
	if u.UserKind != UserKindContact {
		t.Fatalf("UserKind = %q, want contact", u.UserKind)
	}
	if u.DisplayName != "NewName" {
		t.Fatalf("DisplayName = %q", u.DisplayName)
	}
	if !u.LastUpdated.Equal(at) {
		t.Fatalf("LastUpdated = %v, want %v", u.LastUpdated, at)
	}
	if u.FirstSeenAt == nil || !u.FirstSeenAt.Equal(at) {
		t.Fatalf("FirstSeenAt = %v, want %v", u.FirstSeenAt, at)
	}
	if u.LastContactAt == nil || !u.LastContactAt.Equal(at) {
		t.Fatalf("LastContactAt = %v, want %v", u.LastContactAt, at)
	}
}

func TestUserCache_MergeFromGetUserAPI_nonFriend_doesNotDemoteFriend(t *testing.T) {
	now := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	u := &UserCache{VRCUserID: "x", UserKind: UserKindFriend, DisplayName: "F", IsFavorite: true}
	api := &UserCache{DisplayName: "API", Bio: "bio", Status: "offline"}
	u.MergeFromGetUserAPI(false, api, now)
	if u.UserKind != UserKindFriend {
		t.Fatalf("UserKind = %q, want friend", u.UserKind)
	}
	if u.Bio != "bio" || u.DisplayName != "API" {
		t.Fatalf("fields: %+v", u)
	}
}

func TestUserCache_MergeFromGetUserAPI_nonFriend_contact(t *testing.T) {
	now := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	u := &UserCache{VRCUserID: "x", UserKind: UserKindContact, DisplayName: "C"}
	api := &UserCache{DisplayName: "API", Username: "u"}
	u.MergeFromGetUserAPI(false, api, now)
	if u.UserKind != UserKindContact {
		t.Fatalf("UserKind = %q", u.UserKind)
	}
	if u.Username != "u" {
		t.Fatalf("Username = %q", u.Username)
	}
}

func TestUserCache_MergeFromGetUserAPI_friend_upgrades(t *testing.T) {
	now := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	u := &UserCache{VRCUserID: "x", UserKind: UserKindContact}
	api := &UserCache{DisplayName: "Pal", Status: "active", Bio: "b"}
	u.MergeFromGetUserAPI(true, api, now)
	if u.UserKind != UserKindFriend {
		t.Fatalf("UserKind = %q, want friend", u.UserKind)
	}
	if u.Bio != "b" {
		t.Fatalf("Bio not merged")
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
