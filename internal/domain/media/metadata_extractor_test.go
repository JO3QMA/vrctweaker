package media

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultMetadataExtractor_Extract_FromFilename(t *testing.T) {
	extractor := NewDefaultMetadataExtractor()

	tests := []struct {
		name     string
		path     string
		wantID   string
		wantName string
	}{
		{
			name:     "filename with wrld_",
			path:     "/screenshots/VRChat_wrld_abc123def_456.png",
			wantID:   "wrld_abc123def_456",
			wantName: "",
		},
		{
			name:     "filename with wrld and _screenshot suffix",
			path:     "C:\\Pictures\\VRChat_wrld_xyz-789_screenshot.jpg",
			wantID:   "wrld_xyz-789",
			wantName: "",
		},
		{
			name:     "no wrld in filename",
			path:     "/screenshots/VRChat_2025-01-01_12-00-00.001.png",
			wantID:   "",
			wantName: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractor.Extract(tt.path)
			if err != nil {
				t.Fatalf("Extract() err = %v", err)
			}
			if got.WorldID != tt.wantID {
				t.Errorf("Extract() worldID = %q, want %q", got.WorldID, tt.wantID)
			}
			if got.WorldDisplayName != tt.wantName {
				t.Errorf("Extract() worldDisplayName = %q, want %q", got.WorldDisplayName, tt.wantName)
			}
		})
	}
}

func TestDefaultMetadataExtractor_Extract_FromAdjacentFile(t *testing.T) {
	dir := t.TempDir()
	base := "test_screenshot"
	path := filepath.Join(dir, base+".png")
	_ = os.WriteFile(path, []byte("dummy"), 0644)

	extractor := NewDefaultMetadataExtractor()

	// No adjacent file -> empty
	got, err := extractor.Extract(path)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}
	if got.WorldID != "" || got.WorldDisplayName != "" {
		t.Errorf("Extract() without adjacent file: got %+v, want empty", got)
	}

	// Adjacent .txt with wrld_
	txtPath := filepath.Join(dir, base+".txt")
	_ = os.WriteFile(txtPath, []byte("World: wrld_adjacent123\nName: Test World"), 0644)
	got, err = extractor.Extract(path)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}
	if got.WorldID != "wrld_adjacent123" {
		t.Errorf("Extract() worldID = %q, want wrld_adjacent123", got.WorldID)
	}
	if got.WorldDisplayName != "" {
		t.Logf("worldDisplayName extracted: %q (optional)", got.WorldDisplayName)
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
