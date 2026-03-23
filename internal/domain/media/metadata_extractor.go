package media

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

// MetadataExtractor extracts structured metadata from screenshot files.
// Extraction priority: 1) filename/adjacent metafile 2) XMP (JPEG APP1 / PNG iTXt) 3) legacy EXIF/PNG tEXt.
// Non-fatal failures return empty metadata and nil error.
type MetadataExtractor interface {
	Extract(path string) (ScreenshotMetadata, error)
}

// wrldIDRE matches VRChat world IDs (e.g. wrld_abc123, wrld_xyz-456).
var wrldIDRE = regexp.MustCompile(`wrld_[a-zA-Z0-9_-]+`)

// DefaultMetadataExtractor implements MetadataExtractor with filename, adjacent file, and image metadata.
type DefaultMetadataExtractor struct{}

// NewDefaultMetadataExtractor creates a new DefaultMetadataExtractor.
func NewDefaultMetadataExtractor() *DefaultMetadataExtractor {
	return &DefaultMetadataExtractor{}
}

// Extract extracts metadata from the given file path.
func (e *DefaultMetadataExtractor) Extract(path string) (ScreenshotMetadata, error) {
	// Priority 1: filename and adjacent metafile
	if id := extractFromFilename(path); id != "" {
		return ScreenshotMetadata{WorldID: id}, nil
	}
	if id, name := extractFromAdjacentFile(path); id != "" || name != "" {
		return ScreenshotMetadata{WorldID: id, WorldDisplayName: name}, nil
	}

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

	if m.WorldID == "" {
		wid, wname := extractFromImageMetadataBytes(data, ext)
		m.WorldID = wid
		if m.WorldDisplayName == "" {
			m.WorldDisplayName = wname
		}
	}
	if m.TakenAt == nil && (ext == ".jpg" || ext == ".jpeg") {
		m.TakenAt = extractTakenAtFromEXIFData(data)
	}

	return m, nil
}

func extractFromFilename(path string) string {
	base := filepath.Base(path)
	return firstMatch(wrldIDRE, base)
}

func extractFromAdjacentFile(path string) (worldID, worldName string) {
	dir := filepath.Dir(path)
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	candidates := []string{
		filepath.Join(dir, base+".txt"),
		filepath.Join(dir, base+".json"),
	}
	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		content := string(data)
		if id := firstMatch(wrldIDRE, content); id != "" {
			worldName = extractWorldNameFromContent(content, id)
			return id, worldName
		}
	}
	return "", ""
}

func extractFromImageMetadataBytes(data []byte, ext string) (worldID, worldName string) {
	switch ext {
	case ".jpg", ".jpeg":
		return extractFromJPEGData(data)
	case ".png":
		return extractFromPNGData(data)
	}
	return "", ""
}

func extractFromJPEGData(data []byte) (worldID, worldName string) {
	x, err := exif.Decode(bytes.NewReader(data))
	if err != nil {
		return "", ""
	}
	tag, err := x.Get(exif.ImageDescription)
	if err != nil {
		return "", ""
	}
	s, err := tag.StringVal()
	if err != nil {
		return "", ""
	}
	id := firstMatch(wrldIDRE, s)
	if id == "" {
		return "", ""
	}
	return id, extractWorldNameFromContent(s, id)
}

func extractFromPNGData(data []byte) (worldID, worldName string) {
	texts := readPNGTextChunks(data)
	for _, s := range texts {
		id := firstMatch(wrldIDRE, s)
		if id != "" {
			return id, extractWorldNameFromContent(s, id)
		}
	}
	return "", ""
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

// readPNGTextChunks reads tEXt and iTXt chunk contents from PNG data.
func readPNGTextChunks(data []byte) []string {
	const pngSignature = "\x89PNG\r\n\x1a\n"
	if len(data) < len(pngSignature)+12 {
		return nil
	}
	if !bytes.Equal(data[:8], []byte(pngSignature)) {
		return nil
	}
	var result []string
	pos := 8
	for pos+12 <= len(data) {
		length := int(binary.BigEndian.Uint32(data[pos : pos+4]))
		chunkType := string(data[pos+4 : pos+8])
		pos += 8
		if pos+length+4 > len(data) {
			break
		}
		chunkData := data[pos : pos+length]
		pos += length + 4
		if chunkType == "IEND" {
			break
		}
		if chunkType == "tEXt" {
			text := parseTEXXChunk(chunkData)
			if text != "" {
				result = append(result, text)
			}
		}
		if chunkType == "iTXt" {
			text := parseITXTChunk(chunkData)
			if text != "" {
				result = append(result, text)
			}
		}
	}
	return result
}

func parseTEXXChunk(data []byte) string {
	idx := bytes.IndexByte(data, 0)
	if idx < 0 {
		return string(data)
	}
	return string(data[idx+1:])
}

func parseITXTChunk(data []byte) string {
	// iTXt: keyword\0 compFlag compMethod lang\0 transKeyword\0 text
	// For uncompressed (compFlag=0), skip keyword, 2 bytes, lang, transKeyword to get text
	if len(data) < 2 {
		return ""
	}
	idx := bytes.IndexByte(data, 0)
	if idx < 0 {
		return string(data)
	}
	if idx+2 >= len(data) {
		return ""
	}
	if data[idx+1] != 0 {
		return "" // compressed, skip
	}
	pos := idx + 2
	// skip language (null-term)
	idx2 := bytes.IndexByte(data[pos:], 0)
	if idx2 < 0 {
		return string(data[pos:])
	}
	pos += idx2 + 1
	// skip translated keyword (null-term)
	idx3 := bytes.IndexByte(data[pos:], 0)
	if idx3 < 0 {
		return string(data[pos:])
	}
	return string(data[pos+idx3+1:])
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

func extractWorldNameFromContent(content, worldID string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, worldID) {
			// Try to extract quoted or bracketed name near world_id
			if idx := strings.Index(line, worldID); idx >= 0 {
				rest := line[idx+len(worldID):]
				rest = strings.TrimSpace(rest)
				for _, sep := range []string{`"`, `'`, `[`, `{`, `:`, `,`} {
					if strings.HasPrefix(rest, sep) {
						rest = strings.TrimPrefix(rest, sep)
						rest = strings.TrimSpace(rest)
						break
					}
				}
				if end := strings.IndexAny(rest, `"'\]}:,`); end > 0 {
					rest = rest[:end]
				}
				rest = strings.TrimSpace(rest)
				if len(rest) > 0 && len(rest) < 200 {
					return rest
				}
			}
		}
	}
	return ""
}
