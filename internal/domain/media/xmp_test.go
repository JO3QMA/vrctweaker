package media

import (
	"testing"
)

func TestParseVRChatXMP_sampleAttributes(t *testing.T) {
	xmp := `<?xpacket begin="" id="w"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description xmlns:vrc="http://vrchat.com/ns/"
  vrc:WorldID="wrld_db637cfb-64f8-4109-977b-6b755482f133"
  vrc:WorldDisplayName="PARA ROOM"
  vrc:AuthorID="usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e"
  xmlns:xmp="http://ns.adobe.com/xap/1.0/"
  xmp:Author="ぶっちゃん！"
  xmp:CreateDate="2026:02:17 00:01:28.5281072+09:00"
/>
</rdf:RDF>
</x:xmpmeta>`
	m := parseVRChatXMP(xmp)
	if m.WorldID != "wrld_db637cfb-64f8-4109-977b-6b755482f133" {
		t.Errorf("WorldID = %q", m.WorldID)
	}
	if m.WorldDisplayName != "PARA ROOM" {
		t.Errorf("WorldDisplayName = %q", m.WorldDisplayName)
	}
	if m.AuthorVRCUserID != "usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e" {
		t.Errorf("AuthorVRCUserID = %q", m.AuthorVRCUserID)
	}
	if m.AuthorDisplayName != "ぶっちゃん！" {
		t.Errorf("AuthorDisplayName = %q", m.AuthorDisplayName)
	}
	if m.TakenAt == nil {
		t.Fatal("TakenAt nil")
	}
	if m.TakenAt.Year() != 2026 || m.TakenAt.Month() != 2 || m.TakenAt.Day() != 17 {
		t.Errorf("TakenAt date = %v", m.TakenAt)
	}
}

func TestParseXMPDate(t *testing.T) {
	s := "2026:02:17 00:01:28+09:00"
	got, ok := parseXMPDate(s)
	if !ok || got == nil {
		t.Fatalf("parseXMPDate(%q) ok=%v", s, ok)
	}
	if got.Year() != 2026 || got.Month() != 2 || got.Day() != 17 {
		t.Errorf("date = %v", got)
	}
}

func TestNormalizeXMPFractionalSeconds(t *testing.T) {
	in := "00:01:28.1234567890+09:00"
	out := normalizeXMPFractionalSeconds(in)
	want := "00:01:28.123456789+09:00"
	if out != want {
		t.Errorf("got %q want %q", out, want)
	}
}

func TestParseITXTKeywordAndText(t *testing.T) {
	// keyword\0 comp method lang\0 trans\0 text
	payload := []byte("XML:com.adobe.xmp\x00\x00\x00en\x00\x00<x>x</x>")
	kw, text := parseITXTKeywordAndText(payload)
	if kw != "XML:com.adobe.xmp" {
		t.Errorf("keyword = %q", kw)
	}
	if text != "<x>x</x>" {
		t.Errorf("text = %q", text)
	}
}

func TestExtractXMPFromJPEG_minimal(t *testing.T) {
	prefix := []byte("http://ns.adobe.com/xap/1.0/\x00")
	xml := []byte("<x>y</x>")
	payload := append(append([]byte{}, prefix...), xml...)
	segLen := 2 + len(payload)
	data := []byte{0xFF, 0xD8, 0xFF, 0xE1, byte(segLen >> 8), byte(segLen & 0xff)}
	data = append(data, payload...)
	if got := extractXMPFromJPEG(data); got != "<x>y</x>" {
		t.Errorf("extractXMPFromJPEG = %q", got)
	}
}
