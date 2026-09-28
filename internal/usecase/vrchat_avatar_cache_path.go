package usecase

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var vrchatAvatarIDPattern = regexp.MustCompile(
	`^avtr_[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
)

const avatarCacheWalkMaxEntries = 8000

// FindLocalAvatarCachePath searches under cacheRoot for a directory whose path contains avatarID.
// Returns ("", nil) when not found or avatarID is invalid.
func FindLocalAvatarCachePath(cacheRoot, avatarID string) (string, error) {
	avatarID = strings.TrimSpace(avatarID)
	if !vrchatAvatarIDPattern.MatchString(avatarID) {
		return "", nil
	}
	root := strings.TrimSpace(cacheRoot)
	if root == "" {
		return "", nil
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return "", nil
	}

	visited := 0
	var found string
	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		visited++
		if visited > avatarCacheWalkMaxEntries {
			return filepath.SkipAll
		}
		if !d.IsDir() {
			return nil
		}
		if strings.Contains(path, avatarID) && len(path) >= len(found) {
			found = path
		}
		return nil
	})
	if walkErr != nil && walkErr != filepath.SkipAll {
		return "", walkErr
	}
	return found, nil
}
