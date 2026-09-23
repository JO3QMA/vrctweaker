package wailsapp

import (
	"testing"

	"vrchat-tweaker/internal/appversion"
)

func TestApp_GetAppVersion(t *testing.T) {
	appversion.SetVersion("0.1.0")
	app := NewApp()
	if got := app.GetAppVersion(); got != "0.1.0" {
		t.Fatalf("GetAppVersion() = %q", got)
	}
}
