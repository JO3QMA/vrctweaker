package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

// cspMiddleware adds a Content-Security-Policy header to every HTTP response served by
// the AssetServer. This reduces the impact of any XSS reaching the Wails IPC bridge.
//
// Policy rationale:
//   - script-src 'self': Vite bundles all scripts as separate files; inline scripts and
//     eval are blocked. This is the main XSS→IPC mitigation.
//   - style-src 'self' 'unsafe-inline': Vue and Element Plus inject <style> elements at
//     runtime; 'unsafe-inline' is required for styles (risk: CSS injection only, not JS).
//   - img-src 'self' data: https:: VRChat thumbnail URLs are HTTPS.
//   - connect-src 'self': VRChat API calls are made from Go, not from frontend fetch().
//   - object-src 'none': Prevents plugin-based XSS vectors.
//   - base-uri 'self': Prevents base tag hijacking.
//
// Dev vs production: during "wails dev" the WebView loads the Vite dev server
// (see wails.json frontend:dev:serverUrl), so HTML is not served by this
// AssetServer and this header does not apply to that document. Wails bridge
// scripts are injected as external /wails/ipc.js and /wails/runtime.js
// (frontend/vite.config.ts), not inline, so script-src 'self' is not an issue
// there. This middleware applies to the embedded frontend/dist responses.
func cspMiddleware(next http.Handler) http.Handler {
	const policy = "default-src 'self'; " +
		"script-src 'self'; " +
		"style-src 'self' 'unsafe-inline'; " +
		"img-src 'self' data: https:; " +
		"font-src 'self' data:; " +
		"connect-src 'self'; " +
		"object-src 'none'; " +
		"base-uri 'self';"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", policy)
		next.ServeHTTP(w, r)
	})
}

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "VRChat Tweaker",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Middleware: cspMiddleware,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.onShutdown,
		Bind: []interface{}{
			app,
		},
		Frameless: true,
		// Security: right-click context menu (inspect element / DevTools) is disabled
		// in production builds by default (EnableDefaultContextMenu defaults to false).
		// In debug builds ("wails dev"), DevTools remain available for development.
		// Do NOT set EnableDefaultContextMenu: true in production.
		EnableDefaultContextMenu: false,
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: false,
		},
	})

	if err != nil {
		log.Fatal("Error:", err.Error())
	}
}
