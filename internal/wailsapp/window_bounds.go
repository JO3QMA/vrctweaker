package wailsapp

// Desktop window minimum size (logical pixels). Used by main.go Wails options.
// Keep frontend/e2e/fixtures/app-window-bounds.ts in sync when changing these values.
const (
	MinWindowWidth  = 768 // half of 1920px FHD width at 125% DPI scaling
	MinWindowHeight = 540 // half of 1080px FHD height
)
