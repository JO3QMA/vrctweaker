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
