package identity

import (
	"strings"
	"time"
)

// PipelineLocationUnknown is stored in UserCache.Location when the friend is online
// (or on web) but VRChat hides instance details (private world, orange/red, empty world).
// REST friend sync may overwrite this with a concrete location later.
const PipelineLocationUnknown = "pipeline:location_unknown"

// PipelineLocationIsHidden reports whether VRChat marks the friend's instance as not visible.
func PipelineLocationIsHidden(worldID, location string) bool {
	if strings.EqualFold(strings.TrimSpace(worldID), "private") {
		return true
	}
	loc := strings.TrimSpace(location)
	if loc == "" {
		return false
	}
	return strings.EqualFold(loc, "private")
}

// MergeFromPipelineFriendOnline applies friend-online presence. When hideLocation is true,
// or location is empty/private, Location is set to PipelineLocationUnknown (online but instance unknown).
func (u *UserCache) MergeFromPipelineFriendOnline(now time.Time, platform, location string, hideLocation bool) {
	if u == nil || u.UserKind == UserKindSelf {
		return
	}
	u.UserKind = UserKindFriend
	u.LastUpdated = now
	if p := strings.TrimSpace(platform); p != "" {
		u.Platform = p
		u.LastPlatform = p
	}
	loc := strings.TrimSpace(location)
	if hideLocation || PipelineLocationIsHidden("", loc) || loc == "" {
		u.Location = PipelineLocationUnknown
		return
	}
	u.Location = loc
}

// MergeFromPipelineFriendLocation applies friend-location. travelingToLocation is folded
// into Location when location is "traveling" and travelingToLocation is non-empty (VRChat convention).
func (u *UserCache) MergeFromPipelineFriendLocation(now time.Time, location, travelingToLocation, worldID string) {
	if u == nil || u.UserKind == UserKindSelf {
		return
	}
	u.UserKind = UserKindFriend
	u.LastUpdated = now
	if PipelineLocationIsHidden(worldID, location) {
		u.Location = PipelineLocationUnknown
		return
	}
	loc := strings.TrimSpace(location)
	travel := strings.TrimSpace(travelingToLocation)
	if loc == "" {
		u.Location = PipelineLocationUnknown
		return
	}
	if loc == "traveling" && travel != "" {
		u.Location = travel
		return
	}
	u.Location = loc
}

// MergeFromPipelineFriendActive applies friend-active (website activity).
// Friends active on the website are not in VRChat; we still clear "offline" so the UI lists them as online.
func (u *UserCache) MergeFromPipelineFriendActive(now time.Time, platform string) {
	if u == nil || u.UserKind == UserKindSelf {
		return
	}
	u.UserKind = UserKindFriend
	u.LastUpdated = now
	if p := strings.TrimSpace(platform); p != "" {
		u.Platform = p
		u.LastPlatform = p
	}
	if s := strings.TrimSpace(u.Status); s == "" || strings.EqualFold(s, "offline") {
		u.Status = "active"
	}
}

// MergeFromPipelineFriendOffline applies friend-offline.
func (u *UserCache) MergeFromPipelineFriendOffline(now time.Time) {
	if u == nil || u.UserKind == UserKindSelf {
		return
	}
	u.UserKind = UserKindFriend
	u.LastUpdated = now
	u.Status = "offline"
	u.Platform = ""
	u.Location = ""
}

// MergeFromPipelineFriendUser merges profile fields from a full user object (friend-add / friend-update).
// snap must be a friend-shaped row; IsFavorite on the receiver is preserved.
func (u *UserCache) MergeFromPipelineFriendUser(snap *UserCache, now time.Time) {
	if u == nil || snap == nil || u.UserKind == UserKindSelf {
		return
	}
	preserveFav := u.IsFavorite
	u.MergeFromAPIFriend(snap)
	u.IsFavorite = preserveFav
	u.LastUpdated = now
}

// DemoteFriendToContactAfterUnfriend clears friend-specific presence after friend-delete.
// The row remains for encounter history and log-derived contacts.
func (u *UserCache) DemoteFriendToContactAfterUnfriend(now time.Time) {
	if u == nil || u.UserKind == UserKindSelf {
		return
	}
	u.UserKind = UserKindContact
	u.IsFavorite = false
	u.LastUpdated = now
	u.Status = ""
	u.Platform = ""
	u.Location = ""
	u.UserState = ""
}

// MergeFromPipelineSelfUserUpdate merges user-update content into the self row only.
func (u *UserCache) MergeFromPipelineSelfUserUpdate(now time.Time, displayName, status, statusDescription, username string,
	avatarThumb, userIcon, profilePicThumb string,
) {
	if u == nil || u.UserKind != UserKindSelf {
		return
	}
	u.LastUpdated = now
	if displayName != "" {
		u.DisplayName = displayName
	}
	if status != "" {
		u.Status = status
	}
	if statusDescription != "" {
		u.StatusDescription = statusDescription
	}
	if username != "" {
		u.Username = username
	}
	if avatarThumb != "" {
		u.AvatarThumbnailURL = avatarThumb
	}
	if userIcon != "" {
		u.UserIconURL = userIcon
	}
	if profilePicThumb != "" {
		u.ProfilePicOverrideThumbnail = profilePicThumb
	}
}

// MergeFromPipelineSelfLocation updates self row location fields from user-location.
func (u *UserCache) MergeFromPipelineSelfLocation(now time.Time, location, travelingToLocation string) {
	if u == nil || u.UserKind != UserKindSelf {
		return
	}
	u.LastUpdated = now
	loc := strings.TrimSpace(location)
	travel := strings.TrimSpace(travelingToLocation)
	if PipelineLocationIsHidden("", loc) {
		u.Location = PipelineLocationUnknown
		return
	}
	if loc == "traveling" && travel != "" {
		u.Location = travel
		return
	}
	if loc != "" {
		u.Location = loc
	}
}
