package media

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"
)

//go:embed testdata/minimal_exif.jpg
var minimalEXIFJPEG []byte

func TestDefaultMetadataExtractor_Extract_JPEGWithXMP(t *testing.T) {
	dir := t.TempDir()
	xmp := `<?xpacket begin="" id="w"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description xmlns:vrc="http://vrchat.com/ns/"
  vrc:WorldID="wrld_db637cfb-64f8-4109-977b-6b755482f133"
  vrc:WorldDisplayName="Test Room"
  vrc:AuthorID="usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e"
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

	extractor := NewDefaultMetadataExtractor()
	got, err := extractor.Extract(path)
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

	extractor := NewDefaultMetadataExtractor()
	got, err := extractor.Extract(path)
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

	extractor := NewDefaultMetadataExtractor()
	got, err := extractor.Extract(path)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}
	if got.TakenAt == nil {
		t.Fatal("expected TakenAt from EXIF fallback")
	}
}

func TestDefaultMetadataExtractor_Extract_MissingFile(t *testing.T) {
	extractor := NewDefaultMetadataExtractor()
	got, err := extractor.Extract(filepath.Join(t.TempDir(), "missing.jpg"))
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
	extractor := NewDefaultMetadataExtractor()
	got, err := extractor.Extract(path)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}
	if got != (ScreenshotMetadata{}) {
		t.Fatalf("got %+v", got)
	}
}
