package wailsapp

import (
	"runtime"
	"testing"

	"vrchat-tweaker/internal/usecase"
)

func TestApp_ytdlpStatusWhenNotInitialized(t *testing.T) {
	a := NewApp()

	maintain, err := a.GetYTDLPMaintainStatus()
	if err != nil {
		t.Fatalf("GetYTDLPMaintainStatus error: %v", err)
	}
	if maintain.Supported || maintain.UnsupportedReason != "notInitialized" {
		t.Fatalf("GetYTDLPMaintainStatus: got %+v", maintain)
	}

	cookie, err := a.GetYTDLPCookieLinkageStatus()
	if err != nil {
		t.Fatalf("GetYTDLPCookieLinkageStatus error: %v", err)
	}
	if cookie.Supported || cookie.UnsupportedReason != "notInitialized" || cookie.SourceKind != usecase.CookieSourceNone {
		t.Fatalf("GetYTDLPCookieLinkageStatus: got %+v", cookie)
	}
}

func TestApp_ytdlpOperationsWhenNotInitialized(t *testing.T) {
	a := NewApp()
	operations := map[string]func() error{
		"AcknowledgeYTDLPToolsReplaceRisk":  a.AcknowledgeYTDLPToolsReplaceRisk,
		"SetYTDLPToolsReplaceMaintain":      func() error { return a.SetYTDLPToolsReplaceMaintain(true) },
		"AcknowledgeYTDLPCookieLinkageRisk": a.AcknowledgeYTDLPCookieLinkageRisk,
		"SetYTDLPCookieLinkageBrowser":      func() error { return a.SetYTDLPCookieLinkageBrowser("chrome") },
		"SetYTDLPCookieLinkageCookiesFile":  func() error { return a.SetYTDLPCookieLinkageCookiesFile("cookies.txt") },
		"DisableYTDLPCookieLinkage":         a.DisableYTDLPCookieLinkage,
	}

	for name, operation := range operations {
		t.Run(name, func(t *testing.T) {
			if err := operation(); err == nil || err.Error() != "notInitialized" {
				t.Fatalf("error: got %v, want notInitialized", err)
			}
		})
	}
}

func TestApp_RuntimeIsWindowsMatchesRuntime(t *testing.T) {
	if got, want := NewApp().RuntimeIsWindows(), runtime.GOOS == "windows"; got != want {
		t.Fatalf("RuntimeIsWindows: got %t, want %t", got, want)
	}
}
