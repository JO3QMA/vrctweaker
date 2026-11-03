package wailsapp

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestApp_GetStartupInitStatus_loadingThenReady(t *testing.T) {
	a := &App{ctx: context.Background()}
	a.initMu.Lock()
	a.initLoading = true
	a.initPhase = "opening"
	a.initMu.Unlock()

	st := a.GetStartupInitStatus()
	if st.Ready || !st.Loading || st.Phase != "opening" {
		t.Fatalf("status = %+v", st)
	}
}

func TestApp_requireReady_notInitialized(t *testing.T) {
	a := &App{ctx: context.Background()}
	if err := a.requireReady(); !errors.Is(err, errAppNotInitialized) {
		t.Fatalf("err = %v", err)
	}
}

func TestApp_InitializeCore_opensDatabase(t *testing.T) {
	dir := t.TempDir()
	prev := os.Getenv("XDG_CONFIG_HOME")
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Cleanup(func() { _ = os.Setenv("XDG_CONFIG_HOME", prev) })

	a := &App{ctx: context.Background()}
	if err := a.initializeCore(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !a.initReady && a.initErr != nil {
		t.Fatalf("initErr = %v", a.initErr)
	}
	if a.settings == nil {
		t.Fatal("expected settings")
	}
	st := a.GetStartupInitStatus()
	if st.DataDir == "" {
		t.Fatal("expected data dir")
	}
	if _, err := os.Stat(filepath.Join(st.DataDir, "vrchat-tweaker.db")); err != nil {
		t.Fatal(err)
	}
	a.teardownCore()
}

func TestApp_OpenAppDataFolder_unavailableWhenNoPath(t *testing.T) {
	a := &App{ctx: context.Background()}
	if err := a.OpenAppDataFolder(); err == nil {
		t.Fatal("expected error")
	}
}
