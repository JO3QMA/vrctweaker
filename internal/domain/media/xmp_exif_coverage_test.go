package media

import (
	"testing"
	"time"
)

func TestExtractTakenAtFromEXIFData_fromFixture(t *testing.T) {
	got := extractTakenAtFromEXIFData(minimalEXIFJPEG)
	if got == nil {
		t.Fatal("expected TakenAt from fixture EXIF")
	}
	if got.Year() != 2004 {
		t.Fatalf("year = %d", got.Year())
	}
}

func TestExtractTakenAtFromEXIFData_invalidJPEG(t *testing.T) {
	if got := extractTakenAtFromEXIFData([]byte{0xFF, 0xD8, 0xFF, 0xD9}); got != nil {
		t.Fatalf("got %v", got)
	}
}

func TestExtractTakenAtFromEXIFData_emptyEXIFPayload(t *testing.T) {
	payload := append([]byte("Exif\x00\x00"), []byte("not-tiff")...)
	data := []byte{0xFF, 0xD8}
	data = appendJPEGAPP1Segment(data, payload)
	data = append(data, 0xFF, 0xD9)
	if got := extractTakenAtFromEXIFData(data); got != nil {
		t.Fatalf("got %v", got)
	}
}

func TestParseXMPDate_unparseableTime(t *testing.T) {
	if _, ok := parseXMPDate("2026:02:17 not-a-time"); ok {
		t.Fatal("expected false for bad time")
	}
}

func TestParseXMPDate_dateOnlyLocal(t *testing.T) {
	got, ok := parseXMPDate("2025:12:25")
	if !ok || got == nil {
		t.Fatal("expected ok")
	}
	if got.Month() != time.December || got.Day() != 25 {
		t.Fatalf("date = %v", got)
	}
}

func TestExtractXMPFromJPEG_stopsAtSOS(t *testing.T) {
	data := []byte{0xFF, 0xD8, 0xFF, 0xDA, 0x00, 0x08, 0xFF, 0xD9}
	if got := extractXMPFromJPEG(data); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestParseITXTKeywordAndText_missingTranslatedKeywordNull(t *testing.T) {
	payload := []byte("XML:com.adobe.xmp\x00\x00en")
	if kw, text := parseITXTKeywordAndText(payload); kw != "XML:com.adobe.xmp" || text != "" {
		t.Fatalf("kw=%q text=%q", kw, text)
	}
}
