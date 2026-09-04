package media

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEmbedEnrichmentMetadata_JPEG_roundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "shot.jpg")
	xmp := `<?xpacket begin="" id="w"?><x:xmpmeta xmlns:x="adobe:ns:meta/"><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"><rdf:Description xmlns:vrc="http://vrchat.com/ns/" vrc:WorldID="wrld_test" xmlns:xmp="http://ns.adobe.com/xap/1.0/" xmp:CreateDate="2026:02:17 12:00:00+09:00"/></rdf:RDF></x:xmpmeta>`
	data := buildJPEGWithXMPPayload(xmp)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	fields := EnrichmentFields{
		InstanceID: "wrld_test:1~hidden()~region(jp)",
		Participants: []EnrichmentParticipant{
			{VRCUserID: "usr_a", DisplayName: "Alice"},
		},
	}
	if err := EmbedEnrichmentMetadata(path, fields); err != nil {
		t.Fatal(err)
	}
	got, err := ReadEnrichmentFieldsFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.InstanceID != fields.InstanceID {
		t.Fatalf("instance = %q", got.InstanceID)
	}
	if len(got.Participants) != 1 || got.Participants[0].DisplayName != "Alice" {
		t.Fatalf("participants = %+v", got.Participants)
	}
	meta, _ := Extract(path)
	if meta.WorldID != "wrld_test" {
		t.Fatalf("world id changed: %q", meta.WorldID)
	}
}

func TestEmbedEnrichmentMetadata_PNG_roundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "shot.png")
	xmp := `<x:xmpmeta><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"><rdf:Description xmlns:vrc="http://vrchat.com/ns/" vrc:WorldID="wrld_png"/></rdf:RDF></x:xmpmeta>`
	if err := os.WriteFile(path, buildPNGWithITXtXMP(xmp), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EmbedEnrichmentMetadata(path, EnrichmentFields{InstanceID: "wrld_png:42"}); err != nil {
		t.Fatal(err)
	}
	got, err := ReadEnrichmentFieldsFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.InstanceID != "wrld_png:42" {
		t.Fatalf("instance = %q", got.InstanceID)
	}
}

func TestParseEnrichmentXMP(t *testing.T) {
	xmp := `<rdf:Description vrctweaker:InstanceID="inst:1"><vrctweaker:Participants>[{"vrcUserId":"usr_1","displayName":"A"}]</vrctweaker:Participants>`
	got := ParseEnrichmentXMP(xmp)
	if got.InstanceID != "inst:1" || len(got.Participants) != 1 {
		t.Fatalf("got %+v", got)
	}
}
