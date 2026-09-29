package usecase

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var vrchatAvatarIDPattern = regexp.MustCompile(
	`^avtr_[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
)

const avatarCacheWalkMaxEntries = 8000

// ErrAvatarCacheWalkTruncated is returned when the cache directory walk hits the entry limit
// before locating the avatar. Callers may treat this like "path not found" in the UI (empty path).
var ErrAvatarCacheWalkTruncated = errors.New("avatar cache walk truncated at entry limit")

// FindLocalAvatarCachePath searches under cacheRoot for a directory whose base name equals avatarID.
// Returns ("", nil) when not found or avatarID is invalid.
// Returns ("", ErrAvatarCacheWalkTruncated) when the walk limit is exceeded without a match.
func FindLocalAvatarCachePath(cacheRoot, avatarID string) (string, error) {
	return findLocalAvatarCachePathWithLimit(cacheRoot, avatarID, avatarCacheWalkMaxEntries)
}

func findLocalAvatarCachePathWithLimit(cacheRoot, avatarID string, maxEntries int) (string, error) {
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
	truncated := false
	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		visited++
		if visited > maxEntries {
			truncated = true
			return filepath.SkipAll
		}
		if !d.IsDir() {
			return nil
		}
		if filepath.Base(path) == avatarID {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if walkErr != nil && walkErr != filepath.SkipAll {
		return "", walkErr
	}
	if truncated && found == "" {
		return "", ErrAvatarCacheWalkTruncated
	}
	return found, nil
}
