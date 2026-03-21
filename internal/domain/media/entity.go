package media

import "time"

// Screenshot represents a VRChat screenshot with extracted metadata.
type Screenshot struct {
	ID            string
	FilePath      string
	WorldID       string
	WorldName     string
	TakenAt       *time.Time
	FileSizeBytes *int64 // nil if unknown (legacy rows)
}

// ScreenshotThumbnail is persisted JPEG thumbnail bytes and source file stat for cache invalidation.
type ScreenshotThumbnail struct {
	WebpBlob      []byte
	SourceSize    int64
	SourceModUnix int64
}
