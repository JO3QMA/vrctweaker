package media

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultMetadataExtractor_Extract_DoesNotUseFilenameWorldID(t *testing.T) {
	extractor := NewDefaultMetadataExtractor()
	got, err := extractor.Extract("/screenshots/VRChat_wrld_abc123def_456.png")
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}
	if got.WorldID != "" {
		t.Errorf("Extract() worldID = %q, want empty (XMP only)", got.WorldID)
	}
}

func TestDefaultMetadataExtractor_Extract_DoesNotUseAdjacentFile(t *testing.T) {
	dir := t.TempDir()
	base := "test_screenshot"
	path := filepath.Join(dir, base+".png")
	_ = os.WriteFile(path, []byte("dummy"), 0644)

	extractor := NewDefaultMetadataExtractor()

	// Adjacent .txt with wrld_
	txtPath := filepath.Join(dir, base+".txt")
	_ = os.WriteFile(txtPath, []byte("World: wrld_adjacent123\nName: Test World"), 0644)
	got, err := extractor.Extract(path)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}
	if got.WorldID != "" || got.WorldDisplayName != "" {
		t.Errorf("Extract() got %+v, want empty world fields (XMP only)", got)
	}
}

func TestDefaultMetadataExtractor_Extract_NoMatchReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plain_screenshot.png")
	_ = os.WriteFile(path, []byte("fake png"), 0644)

	extractor := NewDefaultMetadataExtractor()
	got, err := extractor.Extract(path)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}
	if got.WorldID != "" || got.WorldDisplayName != "" {
		t.Errorf("Extract() got %+v, want empty world fields", got)
	}
	if got.TakenAt != nil {
		t.Errorf("Extract() takenAt = %v, want nil", got.TakenAt)
	}
}

func TestDefaultMetadataExtractor_Extract_DoesNotUsePNGTextWorldID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "png_text_only.png")
	if err := os.WriteFile(path, buildPNGWithTextChunk("meta", "wrld_from_png_text_only"), 0644); err != nil {
		t.Fatalf("WriteFile() err = %v", err)
	}

	extractor := NewDefaultMetadataExtractor()
	got, err := extractor.Extract(path)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}
	if got.WorldID != "" || got.WorldDisplayName != "" {
		t.Errorf("Extract() got %+v, want empty world fields (XMP only)", got)
	}
}

func buildPNGWithTextChunk(keyword, text string) []byte {
	data := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	chunkText := append([]byte(keyword), 0)
	chunkText = append(chunkText, []byte(text)...)
	data = appendPNGChunk(data, "tEXt", chunkText)
	return appendPNGChunk(data, "IEND", nil)
}

func appendPNGChunk(data []byte, typ string, payload []byte) []byte {
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(payload)))
	data = append(data, lenBuf[:]...)
	data = append(data, []byte(typ)...)
	data = append(data, payload...)
	data = append(data, 0, 0, 0, 0) // CRC is not validated by extractor code.
	return data
}

func TestFirstMatch_WrldID(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{"wrld_abc123", "wrld_abc123"},
		{"wrld_xyz-456", "wrld_xyz-456"},
		{"prefix_wrld_def_789_screenshot", "wrld_def_789"},
		{"no match here", ""},
	}
	for _, tt := range tests {
		m := firstMatch(wrldIDRE, tt.s)
		if m != tt.want {
			t.Errorf("firstMatch(wrldIDRE, %q) = %q, want %q", tt.s, m, tt.want)
		}
	}
}
