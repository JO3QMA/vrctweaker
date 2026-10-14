package tray

// runAsync invokes tray menu handlers without blocking the Win32 message loop.
func runAsync(fn func()) {
	if fn == nil {
		return
	}
	go fn()
}
