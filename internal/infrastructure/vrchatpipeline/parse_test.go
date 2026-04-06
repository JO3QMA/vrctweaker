package vrchatpipeline

import (
	"encoding/json"
	"testing"
)

func TestUnwrapContent_rawObject(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(`{"userId":"usr_1","location":"wrld_x:1"}`)
	out, err := UnwrapContent(raw)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != string(raw) {
		t.Fatalf("got %s", out)
	}
}

func TestUnwrapContent_doubleEncodedObject(t *testing.T) {
	t.Parallel()
	inner := `{"userId":"usr_1"}`
	quoted, err := json.Marshal(inner)
	if err != nil {
		t.Fatal(err)
	}
	out, err := UnwrapContent(quoted)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != inner {
		t.Fatalf("got %s want %s", out, inner)
	}
}

func TestUnwrapContent_plainStringID(t *testing.T) {
	t.Parallel()
	quoted, err := json.Marshal("not_00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatal(err)
	}
	out, err := UnwrapContent(quoted)
	if err != nil {
		t.Fatal(err)
	}
	var s string
	if err := json.Unmarshal(out, &s); err != nil {
		t.Fatal(err)
	}
	if s != "not_00000000-0000-0000-0000-000000000000" {
		t.Fatalf("got %q", s)
	}
}

func TestParseEnvelope(t *testing.T) {
	t.Parallel()
	frame, err := json.Marshal(map[string]string{
		"type":    "friend-offline",
		"content": `{"userId":"usr_abc"}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	typ, payload, err := ParseEnvelope(frame)
	if err != nil {
		t.Fatal(err)
	}
	if typ != "friend-offline" {
		t.Fatalf("type %q", typ)
	}
	var body struct {
		UserID string `json:"userId"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatal(err)
	}
	if body.UserID != "usr_abc" {
		t.Fatalf("userId %q", body.UserID)
	}
}
