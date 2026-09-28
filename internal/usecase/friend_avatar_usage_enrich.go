package usecase

import (
	"strings"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/identity"
)

// EnrichedFriendAvatarUsage is a friend avatar tab row with optional ID and local cache path.
type EnrichedFriendAvatarUsage struct {
	AvatarName     string
	AvatarID       string
	LocalCachePath string
	UseCount       int64
	FirstSeenAt    time.Time
	LastSeenAt     time.Time
}

// AvatarCachePathFinder resolves a local on-disk path for an avatar bundle (tests may inject).
type AvatarCachePathFinder func(cacheRoot, avatarID string) (string, error)

// EnrichFriendAvatarUsages attaches avatar IDs (current equipped match only) and optional cache paths.
func EnrichFriendAvatarUsages(
	list []*activity.FriendAvatarUsageSummary,
	friend *identity.UserCache,
	currentAvatarName string,
	cacheRoot string,
	findPath ...AvatarCachePathFinder,
) []EnrichedFriendAvatarUsage {
	finder := FindLocalAvatarCachePath
	if len(findPath) > 0 && findPath[0] != nil {
		finder = findPath[0]
	}
	currentID := ""
	if friend != nil {
		currentID = strings.TrimSpace(friend.CurrentAvatarID)
	}
	currentName := strings.TrimSpace(currentAvatarName)

	out := make([]EnrichedFriendAvatarUsage, len(list))
	for i, row := range list {
		if row == nil {
			continue
		}
		avatarID := ""
		if currentID != "" && currentName != "" &&
			strings.EqualFold(strings.TrimSpace(row.AvatarName), currentName) {
			avatarID = currentID
		}
		localPath := ""
		if avatarID != "" && strings.TrimSpace(cacheRoot) != "" {
			p, err := finder(cacheRoot, avatarID)
			if err == nil {
				localPath = p
			}
		}
		out[i] = EnrichedFriendAvatarUsage{
			AvatarName:     row.AvatarName,
			AvatarID:       avatarID,
			LocalCachePath: localPath,
			UseCount:       row.UseCount,
			FirstSeenAt:    row.FirstSeenAt,
			LastSeenAt:     row.LastSeenAt,
		}
	}
	return out
}
