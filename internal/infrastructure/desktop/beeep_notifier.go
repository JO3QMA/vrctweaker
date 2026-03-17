package desktop

import (
	"vrchat-tweaker/internal/domain/identity"

	"github.com/gen2brain/beeep"
)

const defaultNotificationTitle = "VRChat Tweaker"

// BeeepNotifier implements identity.Notifier using the beeep library.
// Provides cross-platform desktop notifications (Windows, macOS, Linux).
// Can be swapped for Wails runtime.Notification when available.
var _ identity.Notifier = (*BeeepNotifier)(nil)

// BeeepNotifier sends desktop notifications via beeep.
type BeeepNotifier struct {
	appName string
}

// NewBeeepNotifier creates a notifier with the given app name for notifications.
func NewBeeepNotifier(appName string) *BeeepNotifier {
	if appName == "" {
		appName = defaultNotificationTitle
	}
	return &BeeepNotifier{appName: appName}
}

// NotifyFavoriteOnline sends a desktop notification for a favorite friend coming online.
func (n *BeeepNotifier) NotifyFavoriteOnline(title, message string) error {
	if title == "" {
		title = n.appName
	}
	return beeep.Notify(title, message, "")
}
