package wailsapp

import "context"

// Lifecycle exposes Wails startup/shutdown hooks without binding them to the frontend IPC surface.
type Lifecycle struct {
	app *App
}

// NewLifecycle returns lifecycle hooks for the given App. Do not pass this type to wails.Bind.
func NewLifecycle(app *App) *Lifecycle {
	return &Lifecycle{app: app}
}

// Startup is called from main.OnStartup.
func (l *Lifecycle) Startup(ctx context.Context) {
	l.app.startup(ctx)
}

// Shutdown is called from main.OnShutdown.
func (l *Lifecycle) Shutdown(ctx context.Context) {
	l.app.shutdown(ctx)
}

// BeforeClose is called from main.OnBeforeClose.
func (l *Lifecycle) BeforeClose(ctx context.Context) bool {
	return l.app.beforeClose(ctx)
}
