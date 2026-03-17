package identity

// Notifier sends desktop notifications.
// Implementations may use Wails runtime.Notification or platform-specific APIs.
type Notifier interface {
	// NotifyFavoriteOnline notifies that a favorite friend came online.
	// title: e.g. "VRChat Tweaker"
	// message: e.g. "<displayName> がオンラインになりました"
	NotifyFavoriteOnline(title, message string) error
}
