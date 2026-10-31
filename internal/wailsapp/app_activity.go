package wailsapp

import (
	"os"
	"path/filepath"
	"strings"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/infrastructure/desktop"
	"vrchat-tweaker/internal/platform/paths"
	"vrchat-tweaker/internal/usecase"
)

// activityEncountersChangedEvent is emitted when a new encounter row is written from the log watcher.
const activityEncountersChangedEvent = "activity:encounters-changed"

// activityVideoPlaybackChangedEvent is emitted when video playback history is updated from the log watcher.
const activityVideoPlaybackChangedEvent = "activity:video-playback-changed"

// GetActivityStats returns aggregated play stats for the date range.
func (a *App) GetActivityStats(fromISO, toISO string) (activity.ActivityStats, error) {
	stats, err := a.activity.GetActivityStats(a.ctx, fromISO, toISO)
	if err != nil {
		return activity.ActivityStats{}, err
	}
	if stats == nil {
		return activity.ActivityStats{
			DailyPlaySeconds: []activity.DailyPlaySeconds{},
			TopWorlds:        []activity.TopWorldSummary{},
		}, nil
	}
	if stats.DailyPlaySeconds == nil {
		stats.DailyPlaySeconds = []activity.DailyPlaySeconds{}
	}
	if stats.TopWorlds == nil {
		stats.TopWorlds = []activity.TopWorldSummary{}
	}
	return *stats, nil
}

// Encounters returns user encounters.
func (a *App) Encounters() ([]UserEncounterDTO, error) {
	list, err := a.activity.ListEncountersWithContext(a.ctx, nil)
	if err != nil {
		return nil, err
	}
	return toEncounterDTOsFromContext(list), nil
}

// VideoPlaybackHistory returns video playback attempts for the video tab (newest first).
func (a *App) VideoPlaybackHistory() ([]VideoPlaybackDTO, error) {
	list, err := a.activity.ListVideoPlaybackHistory(a.ctx)
	if err != nil {
		return nil, err
	}
	return toVideoPlaybackDTOs(list), nil
}

// FriendAvatarUsageByVRCUserID returns aggregated avatar usage for a friend (from output_log).
func (a *App) FriendAvatarUsageByVRCUserID(vrcUserID string) ([]FriendAvatarUsageDTO, error) {
	if a.activity == nil {
		return []FriendAvatarUsageDTO{}, nil
	}
	if strings.TrimSpace(vrcUserID) == "" {
		return []FriendAvatarUsageDTO{}, nil
	}
	list, err := a.activity.ListFriendAvatarUsageByVRCUserID(a.ctx, vrcUserID)
	if err != nil {
		return nil, err
	}
	var friend *identity.UserCache
	currentAvatarName := ""
	if a.identity != nil {
		friend, _ = a.identity.CachedUserByVRCUserID(a.ctx, vrcUserID)
		if friend != nil && strings.TrimSpace(friend.CurrentAvatarID) != "" {
			name, nameErr := a.identity.AvatarDisplayName(a.ctx, friend.CurrentAvatarID)
			if nameErr == nil {
				currentAvatarName = name
			}
		}
	}
	cacheRoot := ""
	if a.assetCache != nil {
		resolved, resolveErr := a.assetCache.ResolvePath()
		if resolveErr == nil {
			cacheRoot = resolved
		}
	}
	enriched := usecase.EnrichFriendAvatarUsages(list, friend, currentAvatarName, cacheRoot)
	return toFriendAvatarUsageDTOs(enriched), nil
}

// EncountersByVRCUserID returns encounters for the given VRChat user id. Empty id yields an empty slice.
func (a *App) EncountersByVRCUserID(vrcUserID string) ([]UserEncounterDTO, error) {
	if strings.TrimSpace(vrcUserID) == "" {
		return []UserEncounterDTO{}, nil
	}
	list, err := a.activity.ListEncountersWithContext(a.ctx, &activity.EncounterFilter{VRCUserID: vrcUserID})
	if err != nil {
		return nil, err
	}
	return toEncounterDTOsFromContext(list), nil
}

// EncountersByWorldID returns encounters in the given world. Empty id yields an empty slice.
func (a *App) EncountersByWorldID(worldID string) ([]UserEncounterDTO, error) {
	if strings.TrimSpace(worldID) == "" {
		return []UserEncounterDTO{}, nil
	}
	list, err := a.activity.ListEncountersWithContext(a.ctx, &activity.EncounterFilter{WorldID: worldID})
	if err != nil {
		return nil, err
	}
	return toEncounterDTOsFromContext(list), nil
}

// OpenVRChatLogFolder opens the configured output_log directory (or default VRChat log dir) in the file manager.
func (a *App) OpenVRChatLogFolder() error {
	p, err := a.settings.GetOutputLogPath(a.ctx)
	if err != nil {
		return err
	}
	var dir string
	if strings.TrimSpace(p) != "" {
		abs, err := filepath.Abs(filepath.Clean(p))
		if err != nil {
			return err
		}
		info, err := os.Stat(abs)
		if err != nil {
			return err
		}
		if info.IsDir() {
			dir = abs
		} else {
			dir = filepath.Dir(abs)
		}
	} else {
		dir = paths.DefaultVRChatOutputLogDir()
	}
	return desktop.OpenFolderInFileManager(dir)
}

// RotateEncounters runs the retention cleanup.
func (a *App) RotateEncounters() (int64, error) {
	return a.activity.RotateEncounters(a.ctx)
}
