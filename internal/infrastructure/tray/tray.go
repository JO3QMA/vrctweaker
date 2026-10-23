package tray

// Config holds tray icon and menu labels.
type Config struct {
	Tooltip       string
	IconPath      string
	MenuShowLabel string
	MenuQuitLabel string
	OnShow        func()
	OnQuit        func()
}

// Manager controls the system tray icon lifecycle.
// On Windows, NewManager starts a process-lifetime tray thread that idles between
// Start/Stop sessions; the desktop app does not tear the manager down on quit.
type Manager interface {
	Supported() bool
	Running() bool
	Start(cfg Config) error
	Stop() error
}

// NewManager returns a platform tray manager.
func NewManager() Manager {
	return newManager()
}
