package wailsapp

import (
	"context"
	"errors"
	"runtime"
	"testing"

	"vrchat-tweaker/internal/usecase"
)

func TestApp_ytdlpStatusWhenNotInitialized(t *testing.T) {
	a := NewApp()

	_, err := a.GetYTDLPMaintainStatus()
	if !errors.Is(err, errAppNotInitialized) {
		t.Fatalf("GetYTDLPMaintainStatus error: %v", err)
	}

	_, err = a.GetYTDLPCookieLinkageStatus()
	if !errors.Is(err, errAppNotInitialized) {
		t.Fatalf("GetYTDLPCookieLinkageStatus error: %v", err)
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
			if err := operation(); !errors.Is(err, errAppNotInitialized) {
				t.Fatalf("error: got %v, want %v", err, errAppNotInitialized)
			}
		})
	}
}

func TestApp_ytdlpStatusWhenUseCaseUnsetButInitReady(t *testing.T) {
	a := NewApp()
	a.setInitReadyForTest(context.Background())
	maintain, err := a.GetYTDLPMaintainStatus()
	if err != nil {
		t.Fatal(err)
	}
	if maintain.Supported || maintain.UnsupportedReason != "notInitialized" {
		t.Fatalf("GetYTDLPMaintainStatus: got %+v", maintain)
	}
	cookie, err := a.GetYTDLPCookieLinkageStatus()
	if err != nil {
		t.Fatal(err)
	}
	if cookie.Supported || cookie.UnsupportedReason != "notInitialized" || cookie.SourceKind != usecase.CookieSourceNone {
		t.Fatalf("GetYTDLPCookieLinkageStatus: got %+v", cookie)
	}
}

func TestApp_RuntimeIsWindowsMatchesRuntime(t *testing.T) {
	if got, want := NewApp().RuntimeIsWindows(), runtime.GOOS == "windows"; got != want {
		t.Fatalf("RuntimeIsWindows: got %t, want %t", got, want)
	}
}
