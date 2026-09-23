package vrchatapi

import (
	"strings"
	"testing"

	"vrchat-tweaker/internal/appversion"
)

func TestUserAgentFormat(t *testing.T) {
	appversion.SetVersion("9.8.7")
	ua := UserAgent()
	if !strings.HasPrefix(ua, "VRChat Tweaker/") {
		t.Fatalf("prefix: %q", ua)
	}
	if !strings.Contains(ua, "github.com/JO3QMA/vrctweaker/issues") {
		t.Fatalf("issues URL: %q", ua)
	}
}
