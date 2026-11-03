package wailsapp

import (
	"context"
	"os"
	"testing"
)

func TestEmitStartupStatusChanged_usesLifecycleContextNotCallContext(t *testing.T) {
	var emissions []StartupInitStatus
	startupStatusChangedTestHook = func(st StartupInitStatus) {
		emissions = append(emissions, st)
	}
	t.Cleanup(func() { startupStatusChangedTestHook = nil })

	lifecycle := context.Background()
	a := &App{wailsLifecycleCtx: lifecycle, ctx: lifecycle}
	a.initMu.Lock()
	a.initLoading = true
	a.initPhase = "dedupe"
	a.initMu.Unlock()

	a.emitStartupStatusChanged()

	if len(emissions) != 1 {
		t.Fatalf("emissions = %d", len(emissions))
	}
	if emissions[0].Phase != "dedupe" {
		t.Fatalf("phase = %q", emissions[0].Phase)
	}
}

func TestInitializeCore_emitsStartupStatusDuringInitialRun(t *testing.T) {
	dir := t.TempDir()
	prev := os.Getenv("XDG_CONFIG_HOME")
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Cleanup(func() { _ = os.Setenv("XDG_CONFIG_HOME", prev) })

	var emissions []StartupInitStatus
	startupStatusChangedTestHook = func(st StartupInitStatus) {
		emissions = append(emissions, st)
	}
	t.Cleanup(func() { startupStatusChangedTestHook = nil })

	lifecycle := context.Background()
	initCtx, cancel := context.WithCancel(lifecycle)
	defer cancel()

	a := &App{ctx: lifecycle}
	if err := a.initializeCore(initCtx); err != nil {
		t.Fatal(err)
	}

	if len(emissions) == 0 {
		t.Fatal("expected startup status emissions during initializeCore")
	}
	found := false
	for _, st := range emissions {
		if st.Phase == "opening" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("emissions = %+v", emissions)
	}
	a.teardownCore()
}
