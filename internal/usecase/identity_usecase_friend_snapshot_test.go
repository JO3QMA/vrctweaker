package usecase

import (
	"testing"
	"time"

	"vrchat-tweaker/internal/infrastructure/vrchatapi"
)

func TestUserCacheFromFriend_mapsListFriendsResponse(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	f := vrchatapi.Friend{
		ID:                             "usr_x",
		DisplayName:                    "DN",
		Status:                         "active",
		Username:                       "uname",
		StatusDescription:              "busy",
		UserState:                      "offline",
		CurrentAvatarThumbnailImageURL: "https://t",
		UserIcon:                       "https://i",
		ProfilePicOverrideThumbnail:    "https://ppt",
		Bio:                            "hello",
		BioLinks:                       []string{"https://link"},
		CurrentAvatarImageURL:          "https://full",
		CurrentAvatarTags:              []string{"a", "b"},
		DeveloperType:                  "none",
		FriendKey:                      "fk",
		ImageURL:                       "https://im",
		LastPlatform:                   "win",
		Location:                       "loc",
		LastLogin:                      "2019-08-24T14:15:22Z",
		LastActivity:                   "2019-08-24T15:15:22Z",
		LastMobile:                     "2019-08-24T16:15:22Z",
		Platform:                       "standalonewindows",
		ProfilePicOverride:             "https://ppo",
		Tags:                           []string{"system_trust_basic"},
	}
	u := userCacheFromFriend(f, true, now)
	if u.VRCUserID != "usr_x" || !u.IsFavorite || u.Bio != "hello" {
		t.Fatalf("base: %+v", u)
	}
	if u.BioLinksJSON != `["https://link"]` || u.CurrentAvatarTagsJSON != `["a","b"]` {
		t.Fatalf("json slices: bioLinks=%q avatarTags=%q", u.BioLinksJSON, u.CurrentAvatarTagsJSON)
	}
	if u.Location != "loc" || u.DeveloperType != "none" {
		t.Fatalf("loc/dev: %+v", u)
	}
}
