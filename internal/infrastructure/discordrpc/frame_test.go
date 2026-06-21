package discordrpc

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"
)

func TestParseVoiceSettingsMute(t *testing.T) {
	muted, ok := parseVoiceSettingsMute(json.RawMessage(`{"mute":true}`))
	if !ok || !muted {
		t.Fatalf("mute=true: ok=%v muted=%v", ok, muted)
	}
}

func TestWriteReadFrame_roundTrip(t *testing.T) {
	var buf bytes.Buffer
	payload := map[string]any{"cmd": "PING", "nonce": "1"}
	if err := writeFrame(&buf, opFrame, payload); err != nil {
		t.Fatal(err)
	}
	op, raw, err := readFrame(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if op != opFrame {
		t.Fatalf("opcode %d", op)
	}
	var msg frameMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Cmd != "PING" {
		t.Fatalf("cmd %q", msg.Cmd)
	}
}

func TestReadFrame_emptyBody(t *testing.T) {
	var buf bytes.Buffer
	header := make([]byte, 8)
	binary.LittleEndian.PutUint32(header[4:8], 0)
	_, _ = buf.Write(header)
	op, raw, err := readFrame(&buf)
	if err != nil || op != 0 || len(raw) != 0 {
		t.Fatalf("op=%d raw=%q err=%v", op, raw, err)
	}
}
