package media

import "time"

// ScreenshotMetadata holds extracted metadata from a screenshot file (filename, sidecar, XMP, EXIF, etc.).
// WorldDisplayName and AuthorDisplayName are used at ingest time to upsert world_info and users_cache;
// the screenshots table stores only WorldID and AuthorVRCUserID (see use case / repository).
type ScreenshotMetadata struct {
	WorldID           string
	WorldDisplayName  string
	AuthorVRCUserID   string
	AuthorDisplayName string
	TakenAt           *time.Time
}
