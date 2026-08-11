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

	c.putIfUnchanged("a", true, 0)
	c.putIfUnchanged("b", true, 0)
	c.putIfUnchanged("c", true, 0) // all fresh; evicts arbitrarily down to maxItems
	if len(c.items) != 2 {
		t.Fatalf("items = %d, want %d", len(c.items), c.maxItems)
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
	c.putIfUnchanged("c", true, 0) // a,b are expired -> pruned before adding c
	if len(c.items) != 1 {
		t.Fatalf("items = %d, want 1 (expired pruned)", len(c.items))
	}
	if _, hit, _ := c.get("c"); !hit {
		t.Fatal("expected c cached")
	}
}
