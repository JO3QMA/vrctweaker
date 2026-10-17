package wailsapp

import (
	"errors"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"vrchat-tweaker/internal/platform/safeurl"
)

// OpenExternalURL opens an http(s) URL in the system default browser.
func (a *App) OpenExternalURL(rawURL string) error {
	if a == nil || a.ctx == nil {
		return errors.New("app not initialized")
	}
	safe, err := safeurl.ParseHTTPHTTPS(rawURL)
	if err != nil {
		return err
	}
	runtime.BrowserOpenURL(a.ctx, safe)
	return nil
}
