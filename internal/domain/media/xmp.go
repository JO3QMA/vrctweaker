package media

import (
	"bytes"
	"encoding/binary"
	"regexp"
	"strings"
	"time"
)

const (
	jpegXMPNamespacePrefix = "http://ns.adobe.com/xap/1.0/\x00"
	pngXMPKeyword          = "XML:com.adobe.xmp"
)

// usrIDRE matches VRChat user IDs embedded in XMP (usr_ + hex UUID segments).
var usrIDRE = regexp.MustCompile(`usr_[a-fA-F0-9-]+`)

var (
	reXMPWorldIDAttr          = regexp.MustCompile(`(?:vrc:)?WorldID\s*=\s*"([^"]+)"`)
	reXMPWorldDisplayNameAttr = regexp.MustCompile(`(?:vrc:)?WorldDisplayName\s*=\s*"([^"]*)"`)
	reXMPAuthorIDAttr         = regexp.MustCompile(`(?:vrc:)?AuthorID\s*=\s*"([^"]+)"`)
	reXMPAuthorAttr           = regexp.MustCompile(`(?:xmp:)?Author\s*=\s*"([^"]*)"`)
	reXMPCreateDateAttr       = regexp.MustCompile(`(?:xmp:)?CreateDate\s*=\s*"([^"]+)"`)
	reXMPWorldIDElem          = regexp.MustCompile(`>(wrld_[a-zA-Z0-9_-]+)</[^>]*WorldID`)
	reXMPWorldDisplayElem     = regexp.MustCompile(`(?:vrc:)?WorldDisplayName[^>]*>([^<]*)</`)
	reXMPAuthorIDElem         = regexp.MustCompile(`>(usr_[a-fA-F0-9-]+)</[^>]*AuthorID`)
	reXMPRDFLi                = regexp.MustCompile(`<rdf:li[^>]*>([^<]+)</rdf:li>`)
)

// extractXMPFromJPEG returns raw XMP packet XML from JPEG APP1 segment(s), or empty.
func extractXMPFromJPEG(data []byte) string {
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		return ""
	}
	pos := 2
	var found string
	for pos+4 <= len(data) {
		if data[pos] != 0xFF {
			pos++
			continue
		}
		marker := data[pos+1]
		if marker == 0xD9 || marker == 0xDA {
			break
		}
		if pos+4 > len(data) {
			break
		}
		segLen := int(binary.BigEndian.Uint16(data[pos+2 : pos+4]))
		if segLen < 2 {
			break
		}
		payloadStart := pos + 4
		payloadEnd := payloadStart + segLen - 2
		if payloadEnd > len(data) {
			break
		}
		if marker == 0xE1 {
			payload := data[payloadStart:payloadEnd]
			if len(payload) > len(jpegXMPNamespacePrefix) && bytes.HasPrefix(payload, []byte(jpegXMPNamespacePrefix)) {
				s := string(payload[len(jpegXMPNamespacePrefix):])
				if s != "" {
					found = s
				}
			} else if bytes.Contains(payload, []byte("<x:xmpmeta")) || bytes.Contains(payload, []byte("<?xpacket")) {
				s := string(payload)
				if s != "" {
					found = s
				}
			}
		}
		pos = payloadEnd
	}
	return found
}

// extractXMPFromPNG returns raw XMP XML from PNG iTXt chunk(s), or empty.
func extractXMPFromPNG(data []byte) string {
	const pngSignature = "\x89PNG\r\n\x1a\n"
	if len(data) < len(pngSignature)+12 {
		return ""
	}
	if !bytes.Equal(data[:8], []byte(pngSignature)) {
		return ""
	}
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
		if chunkType != "iTXt" {
			continue
		}
		kw, text := parseITXTKeywordAndText(chunkData)
		if kw == pngXMPKeyword || strings.Contains(text, "<x:xmpmeta") || strings.Contains(text, "<?xpacket") {
			if strings.TrimSpace(text) != "" {
				return text
			}
		}
	}
	return ""
}

// parseITXTKeywordAndText parses PNG iTXt chunk payload: keyword\0 compression lang trans\0 text.
func parseITXTKeywordAndText(data []byte) (keyword string, text string) {
	if len(data) < 3 {
		return "", ""
	}
	idx := bytes.IndexByte(data, 0)
	if idx < 0 {
		return "", ""
	}
	keyword = string(data[:idx])
	if idx+2 >= len(data) {
		return keyword, ""
	}
	if data[idx+1] != 0 {
		return keyword, ""
	}
	pos := idx + 3
	i := bytes.IndexByte(data[pos:], 0)
	if i < 0 {
		return keyword, ""
	}
	pos += i + 1
	i = bytes.IndexByte(data[pos:], 0)
	if i < 0 {
		return keyword, ""
	}
	text = string(data[pos+i+1:])
	return keyword, text
}

// parseVRChatXMP extracts VRChat-oriented fields from an XMP XML string.
func parseVRChatXMP(xmp string) (m ScreenshotMetadata) {
	if xmp == "" {
		return m
	}
	if sub := reXMPWorldIDAttr.FindStringSubmatch(xmp); len(sub) > 1 {
		m.WorldID = firstMatch(wrldIDRE, sub[1])
	}
	if m.WorldID == "" {
		if sub := reXMPWorldIDElem.FindStringSubmatch(xmp); len(sub) > 1 {
			m.WorldID = firstMatch(wrldIDRE, sub[1])
		}
	}
	if m.WorldID == "" {
		m.WorldID = firstMatch(wrldIDRE, xmp)
	}

	if sub := reXMPWorldDisplayNameAttr.FindStringSubmatch(xmp); len(sub) > 1 {
		m.WorldDisplayName = strings.TrimSpace(sub[1])
	}
	if m.WorldDisplayName == "" {
		if sub := reXMPWorldDisplayElem.FindStringSubmatch(xmp); len(sub) > 1 {
			m.WorldDisplayName = strings.TrimSpace(sub[1])
		}
	}

	if sub := reXMPAuthorIDAttr.FindStringSubmatch(xmp); len(sub) > 1 {
		m.AuthorVRCUserID = firstMatchUSR(sub[1])
	}
	if m.AuthorVRCUserID == "" {
		if sub := reXMPAuthorIDElem.FindStringSubmatch(xmp); len(sub) > 1 {
			m.AuthorVRCUserID = firstMatchUSR(sub[1])
		}
	}
	if m.AuthorVRCUserID == "" {
		// Last resort: any usr_ token in XMP (prefer longest match near AuthorID)
		m.AuthorVRCUserID = firstMatchUSR(xmp)
	}

	if sub := reXMPAuthorAttr.FindStringSubmatch(xmp); len(sub) > 1 {
		m.AuthorDisplayName = strings.TrimSpace(sub[1])
	}
	if m.AuthorDisplayName == "" {
		if sub := reXMPRDFLi.FindStringSubmatch(xmp); len(sub) > 1 {
			m.AuthorDisplayName = strings.TrimSpace(sub[1])
		}
	}

	if sub := reXMPCreateDateAttr.FindStringSubmatch(xmp); len(sub) > 1 {
		if t, ok := parseXMPDate(sub[1]); ok {
			m.TakenAt = t
		}
	}

	return m
}

func firstMatchUSR(s string) string {
	m := usrIDRE.FindString(s)
	if m == "" {
		return ""
	}
	return m
}

// parseXMPDate parses XMP date like "2026:02:17 00:01:28.5281072+09:00".
func parseXMPDate(s string) (*time.Time, bool) {
	s = strings.TrimSpace(s)
	if len(s) < 10 {
		return nil, false
	}
	sp := strings.IndexByte(s, ' ')
	if sp < 0 {
		sp = len(s)
	}
	datePart := s[:sp]
	if len(datePart) != 10 || datePart[4] != ':' || datePart[7] != ':' {
		return nil, false
	}
	isoDate := datePart[:4] + "-" + datePart[5:7] + "-" + datePart[8:10]
	timePart := strings.TrimSpace(s[sp:])
	if timePart == "" {
		t, err := time.ParseInLocation("2006-01-02", isoDate, time.Local)
		if err != nil {
			return nil, false
		}
		return &t, true
	}
	// Trim fractional seconds to at most 9 digits for Go
	timePart = normalizeXMPFractionalSeconds(timePart)
	layouts := []string{
		"15:04:05.999999999-07:00",
		"15:04:05.999999999Z07:00",
		"15:04:05-07:00",
		"15:04:05Z07:00",
		"15:04:05.999999999",
		"15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation("2006-01-02 "+layout, isoDate+" "+timePart, time.Local); err == nil {
			return &t, true
		}
	}
	return nil, false
}

func normalizeXMPFractionalSeconds(s string) string {
	// Find . before timezone + or Z
	tzIdx := strings.IndexAny(s, "+-Z")
	if tzIdx < 0 {
		tzIdx = len(s)
	}
	dot := strings.IndexByte(s[:tzIdx], '.')
	if dot < 0 {
		return s
	}
	fracEnd := tzIdx
	frac := s[dot+1 : fracEnd]
	if len(frac) <= 9 {
		return s
	}
	return s[:dot+1+9] + s[fracEnd:]
}
