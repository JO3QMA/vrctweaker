package media

import "time"

// Screenshot represents a VRChat screenshot row.
// WorldName is world_info.display_name when loaded via repository List/Get (JOIN); empty if unknown.
// AuthorDisplayName is resolved from users_cache on read; not persisted on the screenshot row.
type Screenshot struct {
	ID                string
	FilePath          string
	WorldID           string
	WorldName         string // resolved for display (world_info)
	AuthorVRCUserID   string
	AuthorDisplayName string
	TakenAt           *time.Time
	FileSizeBytes     *int64 // nil if unknown (legacy rows)
}

// ScreenshotThumbnail is persisted JPEG thumbnail bytes and source file stat for cache invalidation.
type ScreenshotThumbnail struct {
	JpegBlob      []byte
	SourceSize    int64
	SourceModUnix int64
}
