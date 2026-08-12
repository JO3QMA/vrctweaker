package usecase

import (
	"path/filepath"
	"testing"
	"time"
)

func TestGalleryFileExistsCache_putIfUnchangedRejectsStaleGeneration(t *testing.T) {
	c := newGalleryFileExistsCache()
	c.now = func() time.Time { return time.Unix(1_000_000, 0) }

	_, ok, gen := c.get("p.png")
	if ok {
		t.Fatal("expected miss")
	}
	c.putIfUnchanged("p.png", true, gen)
	if e, hit, _ := c.get("p.png"); !hit || !e {
		t.Fatalf("expected cached exists, got exists=%v hit=%v", e, hit)
	}

	// An ingest completing between get and put must not be overwritten by a
	// stale put carrying the pre-invalidation generation.
	c.invalidatePath("p.png")
	if _, hit, _ := c.get("p.png"); hit {
		t.Fatal("expected miss after invalidate")
	}
	c.putIfUnchanged("p.png", false, gen)
	if _, hit, _ := c.get("p.png"); hit {
		t.Fatal("stale put re-registered a result after invalidation")
	}
}

func TestGalleryFileExistsCache_normalizesPathKeys(t *testing.T) {
	c := newGalleryFileExistsCache()
	c.now = func() time.Time { return time.Unix(1_000_000, 0) }

	c.putIfUnchanged(filepath.Join("a", "..", "p.png"), true, 0)
	if e, hit, _ := c.get("p.png"); !hit || !e {
		t.Fatalf("expected normalized lookup to hit, got exists=%v hit=%v", e, hit)
	}
	c.invalidatePath(filepath.Join("a", "..", "p.png"))
	if _, hit, _ := c.get("p.png"); hit {
		t.Fatal("expected miss after normalized invalidate")
	}
}

func TestGalleryFileExistsCache_evictsWhenOverBound(t *testing.T) {
	c := newGalleryFileExistsCache()
	base := time.Unix(1_000_000, 0)
	c.now = func() time.Time { return base }
	c.maxItems = 2

	// The full sweep only runs at double the bound (4 here); the sweep then
	// evicts arbitrarily down to the steady-state maxItems.
	for _, p := range []string{"a", "b", "c", "d"} {
		c.putIfUnchanged(p, true, 0)
	}
	if len(c.items) != c.maxItems {
		t.Fatalf("items = %d, want %d", len(c.items), c.maxItems)
	}
	// Below the double bound, no pruning happens.
	c.putIfUnchanged("e", true, 0)
	if len(c.items) != c.maxItems+1 {
		t.Fatalf("items = %d, want %d (no sweep under double bound)", len(c.items), c.maxItems+1)
	}
}

func TestGalleryFileExistsCache_prunesExpiredWhenOverBound(t *testing.T) {
	c := newGalleryFileExistsCache()
	base := time.Unix(1_000_000, 0)
	c.now = func() time.Time { return base }
	c.maxItems = 2

	c.putIfUnchanged("a", true, 0)
	c.putIfUnchanged("b", true, 0)
	c.now = func() time.Time { return base.Add(time.Hour) }
	c.putIfUnchanged("c", true, 0)
	c.putIfUnchanged("d", true, 0) // reaching 2*maxItems sweeps expired a,b
	if len(c.items) != c.maxItems {
		t.Fatalf("items = %d, want %d", len(c.items), c.maxItems)
	}
	for _, p := range []string{"a", "b"} {
		if _, hit, _ := c.get(p); hit {
			t.Fatalf("expected %s pruned, still cached", p)
		}
	}
	for _, p := range []string{"c", "d"} {
		if e, hit, _ := c.get(p); !hit || !e {
			t.Fatalf("expected %s cached, got exists=%v hit=%v", p, e, hit)
		}
	}
}

func TestGalleryFileExistsCache_checkCachesOnlyDefinitiveResults(t *testing.T) {
	c := newGalleryFileExistsCache()
	c.now = func() time.Time { return time.Unix(1_000_000, 0) }

	// A transient failure (permission, I/O) is not cached: the next check
	// re-stats and recovers.
	c.stat = func(string) (bool, bool) { return false, false }
	if c.check("p.png") {
		t.Fatal("expected false from transient error")
	}
	c.stat = func(string) (bool, bool) { return true, true }
	if !c.check("p.png") {
		t.Fatal("expected true after transient error cleared")
	}

	// A definitively-missing file IS cached: still missing after a restore.
	c.stat = func(string) (bool, bool) { return false, true }
	if c.check("gone.png") {
		t.Fatal("expected false from not-exist")
	}
	c.stat = func(string) (bool, bool) { return true, true }
	if c.check("gone.png") {
		t.Fatal("expected cached missing within TTL")
	}
}
