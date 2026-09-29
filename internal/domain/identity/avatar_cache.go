package identity

import "time"

// AvatarNameCacheTTL is how long a GET /avatars/{id} display name stays valid.
// It matches UserCacheTTL (30 days). Profile image bytes use a shorter TTL in a later phase.
const AvatarNameCacheTTL = UserCacheTTL

// AvatarNameCache is a persisted avatar id to display name snapshot.
type AvatarNameCache struct {
	AvatarID  string
	Name      string
	FetchedAt time.Time
}
