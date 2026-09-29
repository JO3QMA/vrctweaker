package usecase

import (
	"errors"
	"log"
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
//
// Avatar ID attachment uses a name-match heuristic: only the log row whose avatar name equals the
// friend's current equipped avatar name from the API (case-insensitive) receives currentAvatarId.
// Historical rows and ambiguous cases (same name on different avatars, stale cache names)
// intentionally get no avatarId — the UI keeps name-only rows without inventing IDs.
//
// findPath is optional; when multiple values are passed, only findPath[0] is used.
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

	pathByAvatarID := make(map[string]string)
	out := make([]EnrichedFriendAvatarUsage, 0, len(list))
	for _, row := range list {
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
			if cached, ok := pathByAvatarID[avatarID]; ok {
				localPath = cached
			} else {
				p, err := finder(cacheRoot, avatarID)
				if err == nil {
					localPath = p
				} else if errors.Is(err, ErrAvatarCacheWalkTruncated) {
					log.Printf("[enrich] avatar cache directory walk hit entry limit before match")
				}
				pathByAvatarID[avatarID] = localPath
			}
		}
		out = append(out, EnrichedFriendAvatarUsage{
			AvatarName:     row.AvatarName,
			AvatarID:       avatarID,
			LocalCachePath: localPath,
			UseCount:       row.UseCount,
			FirstSeenAt:    row.FirstSeenAt,
			LastSeenAt:     row.LastSeenAt,
		})
	}
	return out
}
