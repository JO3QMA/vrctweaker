package identity

import "time"

// UserKind classifies a row in users_cache (self, API friend, or log/contact).
type UserKind string

const (
	UserKindSelf    UserKind = "self"
	UserKindFriend  UserKind = "friend"
	UserKindContact UserKind = "contact"
)

// UserCacheTTL is the maximum age of cached VRChat user data before re-fetching from the API (~1 month).
const UserCacheTTL = 30 * 24 * time.Hour

// SettingVRChatFriendsSyncedAt is the app_settings key for last successful friends list sync (RFC3339).
const SettingVRChatFriendsSyncedAt = "vrchat_friends_synced_at"
