package safeurl

import (
	"testing"
)

func TestParseHTTPHTTPS(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantOK  bool
		wantHas string
	}{
		{"https", "https://example.com/path", true, "https://example.com/path"},
		{"http", "http://example.com", true, "http://example.com"},
		{"trim", "  https://a.test  ", true, "https://a.test"},
		{"javascript", "javascript:alert(1)", false, ""},
		{"data", "data:text/html,hi", false, ""},
		{"empty", "", false, ""},
		{"no host", "https://", false, ""},
		{"port only", "http://:80", false, ""},
		{"relative", "/path", false, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseHTTPHTTPS(tc.raw)
			if tc.wantOK {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tc.wantHas != "" && got != tc.wantHas {
					t.Fatalf("got %q want %q", got, tc.wantHas)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error for %q, got %q", tc.raw, got)
			}
		})
	}
}
