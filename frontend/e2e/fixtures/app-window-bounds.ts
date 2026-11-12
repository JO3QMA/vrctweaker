/** Keep in sync with internal/wailsapp/window_bounds.go */
export const APP_MIN_WINDOW_WIDTH = 768;
export const APP_MIN_WINDOW_HEIGHT = 540;

/** Playwright viewport widths aligned with Wails min width and a typical desktop width. */
export const E2E_VIEWPORT_WIDTHS = [1280, APP_MIN_WINDOW_WIDTH] as const;
