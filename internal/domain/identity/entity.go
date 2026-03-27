package identity

import "time"

// UserCache represents a VRChat user row in users_cache (self, friends, and log-derived contacts).
type UserCache struct {
	VRCUserID     string
	DisplayName   string
	Status        string // join me, active, offline, etc.; empty when only seen in logs
	IsFavorite    bool
	LastUpdated   time.Time
	FirstSeenAt   *time.Time
	LastContactAt *time.Time
	UserKind      UserKind
	// SessionFingerprint scopes the self row to the current auth token (hex SHA-256).
	SessionFingerprint string
	// Self-profile fields (GET /auth/user); also stored for user_kind=self rows.
	Username                    string
	StatusDescription           string
	UserState                   string
	AvatarThumbnailURL          string
	UserIconURL                 string
	ProfilePicOverrideThumbnail string
	// List Friends API (GET /auth/user/friends); primarily populated for user_kind=friend.
	Bio                   string
	BioLinksJSON          string
	CurrentAvatarImageURL string
	CurrentAvatarTagsJSON string
	DeveloperType         string
	FriendKey             string
	ImageURL              string
	LastPlatform          string
	Location              string
	LastLogin             string
	LastActivity          string
	LastMobile            string
	Platform              string
	ProfilePicOverride    string
	TagsJSON              string
}

// MergeFromLog merges log-derived contact info without downgrading friend or self rows.
// DisplayName is always updated. LastContactAt moves forward when the log time is newer.
// For contact rows, FirstSeenAt keeps the earliest seen time; for friend/self it is set only if missing (legacy SQL COALESCE).
// LastUpdated is refreshed for contact-only rows; for friend/self it is left unchanged.
func (u *UserCache) MergeFromLog(displayName string, at time.Time) {
	if u == nil {
		return
	}
	u.DisplayName = displayName

	preserveKind := u.UserKind == UserKindFriend || u.UserKind == UserKindSelf
	if !preserveKind {
		u.UserKind = UserKindContact
		u.LastUpdated = at
	}

	if u.LastContactAt == nil || at.After(*u.LastContactAt) {
		t := at
		u.LastContactAt = &t
	}

	if preserveKind {
		if u.FirstSeenAt == nil {
			t := at
			u.FirstSeenAt = &t
		}
		return
	}
	if u.FirstSeenAt == nil || at.Before(*u.FirstSeenAt) {
		t := at
		u.FirstSeenAt = &t
	}
}

// MergeFromAPIFriend merges a friends-list API snapshot into this row.
// Self rows are never modified. Other kinds become friend; IsFavorite comes from the snapshot (caller sets from existing row).
func (u *UserCache) MergeFromAPIFriend(apiUser *UserCache) {
	if u == nil || apiUser == nil {
		return
	}
	if u.UserKind == UserKindSelf {
		return
	}
	u.UserKind = UserKindFriend
	u.DisplayName = apiUser.DisplayName
	u.Status = apiUser.Status
	u.IsFavorite = apiUser.IsFavorite
	u.LastUpdated = apiUser.LastUpdated
	u.Username = apiUser.Username
	u.StatusDescription = apiUser.StatusDescription
	u.UserState = apiUser.UserState
	u.AvatarThumbnailURL = apiUser.AvatarThumbnailURL
	u.UserIconURL = apiUser.UserIconURL
	u.ProfilePicOverrideThumbnail = apiUser.ProfilePicOverrideThumbnail
	u.Bio = apiUser.Bio
	u.BioLinksJSON = apiUser.BioLinksJSON
	u.CurrentAvatarImageURL = apiUser.CurrentAvatarImageURL
	u.CurrentAvatarTagsJSON = apiUser.CurrentAvatarTagsJSON
	u.DeveloperType = apiUser.DeveloperType
	u.FriendKey = apiUser.FriendKey
	u.ImageURL = apiUser.ImageURL
	u.LastPlatform = apiUser.LastPlatform
	u.Location = apiUser.Location
	u.LastLogin = apiUser.LastLogin
	u.LastActivity = apiUser.LastActivity
	u.LastMobile = apiUser.LastMobile
	u.Platform = apiUser.Platform
	u.ProfilePicOverride = apiUser.ProfilePicOverride
	u.TagsJSON = apiUser.TagsJSON
}
