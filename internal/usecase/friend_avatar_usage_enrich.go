package usecase

import (
	"strings"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/identity"
)

// EnrichedFriendAvatarUsage is a friend avatar tab row with optional avatar ID.
type EnrichedFriendAvatarUsage struct {
	AvatarName  string
	AvatarID    string
	UseCount    int64
	FirstSeenAt time.Time
	LastSeenAt  time.Time
}

// EnrichFriendAvatarUsages attaches avatar IDs (current equipped match only).
//
// Avatar ID attachment uses a name-match heuristic: only the log row whose avatar name equals the
// friend's current equipped avatar name from the API (case-insensitive) receives currentAvatarId.
// Historical rows and ambiguous cases (same name on different avatars, stale cache names)
// intentionally get no avatarId — the UI keeps name-only rows without inventing IDs.
func EnrichFriendAvatarUsages(
	list []*activity.FriendAvatarUsageSummary,
	friend *identity.UserCache,
	currentAvatarName string,
) []EnrichedFriendAvatarUsage {
	currentID := ""
	if friend != nil {
		currentID = strings.TrimSpace(friend.CurrentAvatarID)
	}
	currentName := strings.TrimSpace(currentAvatarName)

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
		out = append(out, EnrichedFriendAvatarUsage{
			AvatarName:  row.AvatarName,
			AvatarID:    avatarID,
			UseCount:    row.UseCount,
			FirstSeenAt: row.FirstSeenAt,
			LastSeenAt:  row.LastSeenAt,
		})
	}
	return out
}
