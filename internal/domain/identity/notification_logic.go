package identity

import "strings"

// StatusOffline is the VRChat status string for offline.
const StatusOffline = "offline"

// IsOffline returns true if status indicates the user is offline.
// VRChat uses "offline" for offline; "active", "join me", "ask me", "busy" etc. are online.
func IsOffline(status string) bool {
	return strings.TrimSpace(strings.ToLower(status)) == StatusOffline
}

// DetectFavoriteOnlineTransitions finds favorites that transitioned from offline to online.
// before: map of vrcUserID -> status before refresh (from ListFavorites)
// after:  map of vrcUserID -> UserCache after refresh (only favorites)
// Returns list of friends who were offline and are now online.
func DetectFavoriteOnlineTransitions(
	before map[string]string,
	after map[string]*UserCache,
) []*UserCache {
	var result []*UserCache
	for id, fc := range after {
		prevStatus := before[id]
		if IsOffline(prevStatus) && !IsOffline(fc.Status) {
			result = append(result, fc)
		}
	}
	return result
}
