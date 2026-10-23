//go:build windows

package w32systray

import (
	"testing"
	"unicode/utf16"
)

func TestCopyUTF16Field_truncatesWithNullTerminator(t *testing.T) {
	dst := make([]uint16, 4)
	long := "abcdef"
	copyUTF16Field(dst, long)
	if dst[3] != 0 {
		t.Fatalf("expected truncated null at index 3, got %d", dst[3])
	}
	got := string(utf16.Decode(dst[:4]))
	if got != "abc\x00" {
		t.Fatalf("unexpected decode: %q", got)
	}
}
