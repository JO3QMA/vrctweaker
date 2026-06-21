package discordrpc

import "testing"

func TestParseEnvLine(t *testing.T) {
	tests := []struct {
		line   string
		key    string
		value  string
		wantOK bool
	}{
		{"VRCTWEAKER_DISCORD_CLIENT_ID=abc", "VRCTWEAKER_DISCORD_CLIENT_ID", "abc", true},
		{`export VRCTWEAKER_DISCORD_CLIENT_ID="123"`, "VRCTWEAKER_DISCORD_CLIENT_ID", "123", true},
		{"# comment", "", "", false},
		{"", "", "", false},
		{"NOT_DISCORD=1", "NOT_DISCORD", "1", true},
	}
	for _, tt := range tests {
		key, value, ok := parseEnvLine(tt.line)
		if ok != tt.wantOK {
			t.Fatalf("line %q: ok=%v want %v", tt.line, ok, tt.wantOK)
		}
		if !tt.wantOK {
			continue
		}
		if key != tt.key || value != tt.value {
			t.Fatalf("line %q: got %q=%q want %q=%q", tt.line, key, value, tt.key, tt.value)
		}
	}
}

func TestClientIDConfigured(t *testing.T) {
	old := ClientID
	t.Cleanup(func() { ClientID = old })

	ClientID = "000000000000000000"
	if clientIDConfigured() {
		t.Fatal("placeholder should not be configured")
	}
	ClientID = "123456789012345678"
	if !clientIDConfigured() {
		t.Fatal("expected configured")
	}
}
