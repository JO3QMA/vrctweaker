package media

import (
	_ "embed"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"vrchat-tweaker/internal/testvrc"
)

//go:embed testdata/minimal_exif.jpg
var minimalEXIFJPEG []byte

func TestDefaultMetadataExtractor_Extract_DoesNotUseFilenameWorldID(t *testing.T) {
	got, err := Extract("/screenshots/VRChat_wrld_abc123def_456.png")
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

	// Adjacent .txt with wrld_
	txtPath := filepath.Join(dir, base+".txt")
	_ = os.WriteFile(txtPath, []byte("World: wrld_adjacent123\nName: Test World"), 0644)
	got, err := Extract(path)
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

	got, err := Extract(path)
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

	got, err := Extract(path)
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

func TestDefaultMetadataExtractor_Extract_JPEGWithXMP(t *testing.T) {
	dir := t.TempDir()
	xmp := `<?xpacket begin="" id="w"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description xmlns:vrc="http://vrchat.com/ns/"
  vrc:WorldID="wrld_db637cfb-64f8-4109-977b-6b755482f133"
  vrc:WorldDisplayName="Test Room"
  vrc:AuthorID="` + testvrc.PlayerUserID + `"
  xmlns:xmp="http://ns.adobe.com/xap/1.0/"
  xmp:Author="Tester"
  xmp:CreateDate="2026:02:17 00:01:28+09:00"
/>
</rdf:RDF>
</x:xmpmeta>`
	path := filepath.Join(dir, "shot.jpg")
	if err := os.WriteFile(path, buildJPEGWithXMPPayload(xmp), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Extract(path)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}
	if got.WorldID != "wrld_db637cfb-64f8-4109-977b-6b755482f133" {
		t.Errorf("WorldID = %q", got.WorldID)
	}
	if got.TakenAt == nil {
		t.Fatal("TakenAt nil")
	}
}

func TestDefaultMetadataExtractor_Extract_PNGWithXMP(t *testing.T) {
	dir := t.TempDir()
	xmp := `<x:xmpmeta xmlns:x="adobe:ns:meta/">
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description xmlns:vrc="http://vrchat.com/ns/"
  vrc:WorldID="wrld_png123-4567-8901-2345-678901234567"
  vrc:WorldDisplayName="PNG Room"
/>
</rdf:RDF>
</x:xmpmeta>`
	path := filepath.Join(dir, "shot.png")
	if err := os.WriteFile(path, buildPNGWithITXtXMP(xmp), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Extract(path)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}
	if got.WorldID != "wrld_png123-4567-8901-2345-678901234567" {
		t.Errorf("WorldID = %q", got.WorldID)
	}
	if got.WorldDisplayName != "PNG Room" {
		t.Errorf("WorldDisplayName = %q", got.WorldDisplayName)
	}
}

func TestDefaultMetadataExtractor_Extract_JPEGEXIFFallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exif.jpg")
	if err := os.WriteFile(path, minimalEXIFJPEG, 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Extract(path)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}
	if got.TakenAt == nil {
		t.Fatal("expected TakenAt from EXIF fallback")
	}
}

func TestDefaultMetadataExtractor_Extract_MissingFile(t *testing.T) {
	got, err := Extract(filepath.Join(t.TempDir(), "missing.jpg"))
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}
	if got != (ScreenshotMetadata{}) {
		t.Fatalf("got %+v", got)
	}
}

func TestDefaultMetadataExtractor_Extract_UnsupportedExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := Extract(path)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}
	if got != (ScreenshotMetadata{}) {
		t.Fatalf("got %+v", got)
	}
}
