package media

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

// MetadataExtractor extracts structured metadata from screenshot files.
// Extraction policy:
//   - World/author fields are read from XMP only (JPEG APP1 / PNG iTXt).
//   - TakenAt is read from XMP first, then JPEG EXIF DateTimeOriginal/DateTime as fallback.
//
// Non-fatal failures return empty metadata and nil error.
type MetadataExtractor interface {
	Extract(path string) (ScreenshotMetadata, error)
}

// wrldIDRE matches VRChat world IDs (e.g. wrld_abc123, wrld_xyz-456).
var wrldIDRE = regexp.MustCompile(`wrld_[a-zA-Z0-9_-]+`)

// DefaultMetadataExtractor implements MetadataExtractor using screenshot embedded metadata.
type DefaultMetadataExtractor struct{}

// NewDefaultMetadataExtractor creates a new DefaultMetadataExtractor.
func NewDefaultMetadataExtractor() *DefaultMetadataExtractor {
	return &DefaultMetadataExtractor{}
}

// Extract extracts metadata from the given file path.
func (e *DefaultMetadataExtractor) Extract(path string) (ScreenshotMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ScreenshotMetadata{}, nil
	}
	ext := strings.ToLower(filepath.Ext(path))

	var m ScreenshotMetadata
	switch ext {
	case ".jpg", ".jpeg":
		if x := extractXMPFromJPEG(data); x != "" {
			m = parseVRChatXMP(x)
		}
	case ".png":
		if x := extractXMPFromPNG(data); x != "" {
			m = parseVRChatXMP(x)
		}
	}

	if m.TakenAt == nil && (ext == ".jpg" || ext == ".jpeg") {
		m.TakenAt = extractTakenAtFromEXIFData(data)
	}

	return m, nil
}

func extractTakenAtFromEXIFData(data []byte) *time.Time {
	x, err := exif.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	tag, err := x.Get(exif.DateTimeOriginal)
	if err != nil {
		tag, err = x.Get(exif.DateTime)
		if err != nil {
			return nil
		}
	}
	s, err := tag.StringVal()
	if err != nil {
		return nil
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	// EXIF datetime: "2006:01:02 15:04:05"
	t, err := time.ParseInLocation("2006:01:02 15:04:05", s, time.Local)
	if err != nil {
		return nil
	}
	return &t
}

func firstMatch(re *regexp.Regexp, s string) string {
	m := re.FindString(s)
	if m == "" {
		return ""
	}
	// Trim common filename suffixes that may be concatenated (e.g. wrld_xyz_screenshot)
	for _, suffix := range []string{"_screenshot", "_thumb", ".png", ".jpg", ".jpeg"} {
		m = strings.TrimSuffix(m, suffix)
	}
	return m
}
