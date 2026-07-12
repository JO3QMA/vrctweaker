package ytdlpmaintain

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

type fakeSettings struct {
	on bool
}

func (f *fakeSettings) YTDLPToolsReplaceMaintain(context.Context) (bool, error) {
	return f.on, nil
}

type fakeProc struct {
	running bool
}

func (f *fakeProc) VRChatRunning() (bool, error) {
	return f.running, nil
}

type countingReapplier struct {
	n atomic.Int32
}

func (c *countingReapplier) ReapplyIfNeeded(context.Context) error {
	c.n.Add(1)
	return nil
}

type fixedToolsDir struct {
	dir string
}

func (f fixedToolsDir) ToolsDir() (string, error) {
	return f.dir, nil
}

func TestRun_reappliesWhenMaintainAndVRChat(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	settings := &fakeSettings{on: true}
	proc := &fakeProc{running: true}
	reap := &countingReapplier{}

	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, 50*time.Millisecond, settings, proc, reap, fixedToolsDir{dir: dir})
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if reap.n.Load() >= 1 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if reap.n.Load() < 1 {
		t.Fatal("expected at least one reapply")
	}

	// Simulate rollback file write
	before := reap.n.Load()
	if err := os.WriteFile(filepath.Join(dir, "yt-dlp.exe"), []byte("bundled"), 0o644); err != nil {
		t.Fatal(err)
	}
	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if reap.n.Load() > before {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if reap.n.Load() <= before {
		t.Fatal("expected reapply after tools write")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit")
	}
}

func TestRun_skipsWhenMaintainOff(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	reap := &countingReapplier{}
	_ = Run(ctx, 30*time.Millisecond, &fakeSettings{on: false}, &fakeProc{running: true}, reap, fixedToolsDir{dir: dir})
	if reap.n.Load() != 0 {
		t.Fatalf("reapply count %d", reap.n.Load())
	}
}
