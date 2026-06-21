package media

import (
	"encoding/binary"
	"testing"
	"time"
)

func appendJPEGAPP1Segment(data []byte, payload []byte) []byte {
	segLen := 2 + len(payload)
	data = append(data, 0xFF, 0xE1, byte(segLen>>8), byte(segLen&0xff))
	data = append(data, payload...)
	return data
}

func buildJPEGWithXMPPayload(xmp string) []byte {
	data := []byte{0xFF, 0xD8}
	prefix := []byte(jpegXMPNamespacePrefix)
	payload := append(append([]byte{}, prefix...), []byte(xmp)...)
	data = appendJPEGAPP1Segment(data, payload)
	return append(data, 0xFF, 0xD9)
}

func buildPNGWithITXtXMP(xmp string) []byte {
	data := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	chunkText := append([]byte(pngXMPKeyword), 0, 0)
	chunkText = append(chunkText, []byte("en")...)
	chunkText = append(chunkText, 0)
	chunkText = append(chunkText, 0) // translated keyword empty
	chunkText = append(chunkText, []byte(xmp)...)
	data = appendPNGChunk(data, "iTXt", chunkText)
	return appendPNGChunk(data, "IEND", nil)
}

func TestExtractXMPFromJPEG_invalidAndAlternatePaths(t *testing.T) {
	if got := extractXMPFromJPEG([]byte{0x00}); got != "" {
		t.Fatalf("invalid header: got %q", got)
	}
	if got := extractXMPFromJPEG([]byte{0xFF, 0xD8, 0xFF, 0xD9}); got != "" {
		t.Fatalf("empty jpeg: got %q", got)
	}

	xpacket := `<?xpacket begin=""?><x:xmpmeta xmlns:x="adobe:ns:meta/"><rdf:RDF/></x:xmpmeta>`
	data := []byte{0xFF, 0xD8}
	data = appendJPEGAPP1Segment(data, []byte(xpacket))
	data = append(data, 0xFF, 0xD9)
	if got := extractXMPFromJPEG(data); got != xpacket {
		t.Fatalf("xpacket path: got %q", got)
	}
}

func TestExtractXMPFromPNG_itxtChunk(t *testing.T) {
	xmp := `<x:xmpmeta><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"/></x:xmpmeta>`
	data := buildPNGWithITXtXMP(xmp)
	if got := extractXMPFromPNG(data); got != xmp {
		t.Fatalf("extractXMPFromPNG = %q, want %q", got, xmp)
	}
	if got := extractXMPFromPNG([]byte("not png")); got != "" {
		t.Fatalf("non-png: got %q", got)
	}
}

func TestParseITXTKeywordAndText_edgeCases(t *testing.T) {
	if kw, text := parseITXTKeywordAndText([]byte("a")); kw != "" || text != "" {
		t.Fatalf("short payload: kw=%q text=%q", kw, text)
	}
	if kw, text := parseITXTKeywordAndText([]byte("keyonly")); kw != "" || text != "" {
		t.Fatalf("no delimiter: kw=%q text=%q", kw, text)
	}
	if kw, text := parseITXTKeywordAndText([]byte("key\x00\x01")); kw != "key" || text != "" {
		t.Fatalf("compressed: kw=%q text=%q", kw, text)
	}
}

func TestParseVRChatXMP_elementFallbacks(t *testing.T) {
	xmp := `<x:xmpmeta>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<WorldID>wrld_db637cfb-64f8-4109-977b-6b755482f133</WorldID>
<vrc:WorldDisplayName>Elem World</vrc:WorldDisplayName>
<AuthorID>usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e</AuthorID>
<rdf:li>Elem Author</rdf:li>
</rdf:RDF>
</x:xmpmeta>`
	m := parseVRChatXMP(xmp)
	if m.WorldID != "wrld_db637cfb-64f8-4109-977b-6b755482f133" {
		t.Errorf("WorldID = %q", m.WorldID)
	}
	if m.WorldDisplayName != "Elem World" {
		t.Errorf("WorldDisplayName = %q", m.WorldDisplayName)
	}
	if m.AuthorVRCUserID != "usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e" {
		t.Errorf("AuthorVRCUserID = %q", m.AuthorVRCUserID)
	}
	if m.AuthorDisplayName != "Elem Author" {
		t.Errorf("AuthorDisplayName = %q", m.AuthorDisplayName)
	}
}

func TestParseVRChatXMP_worldIDFromBody(t *testing.T) {
	xmp := `metadata wrld_fallback123-4567-8901-2345-678901234567 more text`
	m := parseVRChatXMP(xmp)
	if m.WorldID != "wrld_fallback123-4567-8901-2345-678901234567" {
		t.Errorf("WorldID = %q", m.WorldID)
	}
}

func TestParseVRChatXMP_empty(t *testing.T) {
	m := parseVRChatXMP("")
	if m != (ScreenshotMetadata{}) {
		t.Fatalf("got %+v", m)
	}
}

func TestFirstMatchUSR(t *testing.T) {
	if got := firstMatchUSR("prefix usr_abc-def-123 suffix"); got != "usr_abc-def-123" {
		t.Errorf("got %q", got)
	}
	if got := firstMatchUSR("no id"); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestParseXMPDate_variants(t *testing.T) {
	t.Run("date only", func(t *testing.T) {
		got, ok := parseXMPDate("2026:03:01")
		if !ok || got == nil {
			t.Fatal("expected ok")
		}
		if got.Year() != 2026 || got.Month() != time.March || got.Day() != 1 {
			t.Fatalf("date = %v", got)
		}
	})
	t.Run("invalid date part", func(t *testing.T) {
		if _, ok := parseXMPDate("bad"); ok {
			t.Fatal("expected false")
		}
	})
	t.Run("fractional Z", func(t *testing.T) {
		got, ok := parseXMPDate("2026:02:17 00:01:28.1234567890Z")
		if !ok || got == nil {
			t.Fatal("expected ok")
		}
	})
	t.Run("no fractional", func(t *testing.T) {
		got, ok := parseXMPDate("2026:02:17 00:01:28")
		if !ok || got == nil {
			t.Fatal("expected ok")
		}
	})
}

func TestNormalizeXMPFractionalSeconds_noDot(t *testing.T) {
	in := "00:01:28+09:00"
	if out := normalizeXMPFractionalSeconds(in); out != in {
		t.Errorf("got %q", out)
	}
}

func TestExtractXMPFromJPEG_truncatedSegment(t *testing.T) {
	data := []byte{0xFF, 0xD8, 0xFF, 0xE1, 0x00, 0x01}
	if got := extractXMPFromJPEG(data); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractXMPFromPNG_nonITXtSkipped(t *testing.T) {
	data := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], 0)
	data = append(data, lenBuf[:]...)
	data = append(data, []byte("tEXt")...)
	data = append(data, 0, 0, 0, 0)
	data = appendPNGChunk(data, "IEND", nil)
	if got := extractXMPFromPNG(data); got != "" {
		t.Fatalf("got %q", got)
	}
}
