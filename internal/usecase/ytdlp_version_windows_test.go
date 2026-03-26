//go:build windows

package usecase

import "testing"

func TestLocalYTDLPFileVersionString_nonexistent(t *testing.T) {
	t.Parallel()
	if g := localYTDLPFileVersionString(`C:\path\that\does\not\exist\yt-dlp.exe`); g != "" {
		t.Fatalf("want empty, got %q", g)
	}
}
