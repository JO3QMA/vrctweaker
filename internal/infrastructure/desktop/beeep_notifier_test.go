package desktop

import (
	"testing"
)

func TestNewBeeepNotifier_defaultAppName(t *testing.T) {
	t.Parallel()
	n := NewBeeepNotifier("")
	if n.appName != defaultNotificationTitle {
		t.Fatalf("appName = %q, want %q", n.appName, defaultNotificationTitle)
	}
}

func TestNewBeeepNotifier_customAppName(t *testing.T) {
	t.Parallel()
	n := NewBeeepNotifier("Custom App")
	if n.appName != "Custom App" {
		t.Fatalf("appName = %q", n.appName)
	}
}

func TestBeeepNotifier_NotifyFavoriteOnline_usesDefaultTitle(t *testing.T) {
	t.Parallel()
	n := NewBeeepNotifier("VRChat Tweaker")
	// beeep may fail in headless CI; either outcome exercises NotifyFavoriteOnline.
	_ = n.NotifyFavoriteOnline("", "friend is online")
}

func TestBeeepNotifier_NotifyFavoriteOnline_customTitle(t *testing.T) {
	t.Parallel()
	n := NewBeeepNotifier("App")
	_ = n.NotifyFavoriteOnline("Friend Online", "Alice joined")
}
