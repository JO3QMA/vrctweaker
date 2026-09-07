package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"vrchat-tweaker/internal/infrastructure/singleinstance"
	"vrchat-tweaker/internal/wailsapp"
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

//go:embed build/windows/icon.ico
var trayIconICO []byte

func main() {
	if len(trayIconICO) > 0 {
		wailsapp.SetTrayIconICO(trayIconICO)
	}

	instanceGuard := singleinstance.New()
	acquired, err := instanceGuard.Acquire()
	if err != nil {
		log.Fatal("single instance: ", err)
	}
	if !acquired {
		if notifyErr := instanceGuard.NotifyExisting(); notifyErr != nil {
			log.Println("Another instance is already running:", notifyErr)
		} else {
			log.Println("Another instance is already running; activated existing window")
		}
		os.Exit(0)
	}
	if err := instanceGuard.Start(); err != nil {
		log.Fatal("single instance listener: ", err)
	}
	defer instanceGuard.Release()

	app := wailsapp.NewApp()
	lc := wailsapp.NewLifecycle(app)

	err = wails.Run(&options.App{
		Title:  singleinstance.DefaultWindowTitle,
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Middleware: cspMiddleware,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			lc.Startup(ctx)
			instanceGuard.SetOnActivate(app.ActivateMainWindow)
		},
		OnShutdown: func(ctx context.Context) {
			lc.Shutdown(ctx)
			instanceGuard.Release()
		},
		OnBeforeClose: lc.BeforeClose,
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
