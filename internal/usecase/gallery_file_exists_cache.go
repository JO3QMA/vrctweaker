package usecase

import (
	"path/filepath"
	"sync"
	"time"
)

// galleryFileStatCacheTTL bounds how long a cached file-existence result is
// reused by Gallery listing before the path is re-checked on disk. A short TTL
// keeps external deletions/restorations from staying stale for long while
// avoiding an os.Stat per row on every listing (ADR 0001).
const galleryFileStatCacheTTL = 30 * time.Second

// galleryFileStatCacheMaxItems caps the number of cached paths. Paths that stop
// appearing in listings (e.g. their DB row was removed) are pruned once the
// cache exceeds this bound so long-running use does not grow without limit.
const galleryFileStatCacheMaxItems = 8192

type galleryFileStatEntry struct {
	exists bool
	at     time.Time
}

// galleryFileExistsCache memoizes per-path file-existence results used by
// Gallery listing. Entries expire after ttl; InvalidateAll / InvalidatePath
// drop entries early when the picture folder is synced or a file is ingested.
type galleryFileExistsCache struct {
	mu         sync.Mutex
	ttl        time.Duration
	now        func() time.Time
	generation uint64
	maxItems   int
	items      map[string]galleryFileStatEntry
}

func newGalleryFileExistsCache() *galleryFileExistsCache {
	return &galleryFileExistsCache{
		ttl:      galleryFileStatCacheTTL,
		now:      time.Now,
		maxItems: galleryFileStatCacheMaxItems,
		items:    make(map[string]galleryFileStatEntry),
	}
}

// get returns the cached existence for path when it is still fresh, along with
// the generation observed. The generation must be passed back to
// putIfUnchanged so a result computed before a concurrent invalidation is not
// re-registered afterwards.
func (c *galleryFileExistsCache) get(path string) (exists bool, ok bool, generation uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	path = c.normalize(path)
	e, ok := c.items[path]
	if !ok {
		return false, false, c.generation
	}
	if c.now().Sub(e.at) >= c.ttl {
		delete(c.items, path)
		return false, false, c.generation
	}
	return e.exists, true, c.generation
}

// putIfUnchanged records the existence of path only when no invalidation has
// happened since the caller's get. A put carrying a stale generation is dropped
// so an ingest that completed in between is never overwritten by an older
// (e.g. "missing") result.
func (c *galleryFileExistsCache) putIfUnchanged(path string, exists bool, generation uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	path = c.normalize(path)
	if generation != c.generation {
		return
	}
	c.items[path] = galleryFileStatEntry{exists: exists, at: c.now()}
	c.pruneLocked()
}

// invalidateAll drops every cached entry and bumps the generation so in-flight
// puts are rejected. Called after a folder sync or ingest.
func (c *galleryFileExistsCache) invalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.generation++
	c.items = make(map[string]galleryFileStatEntry)
}

// invalidatePath drops the entry for one path and bumps the generation. Called
// when a file is ingested.
func (c *galleryFileExistsCache) invalidatePath(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.generation++
	delete(c.items, c.normalize(path))
}

// normalize gives all callers one key space so different spellings of the same
// file (relative vs absolute, "a/../b", …) hit a single entry.
func (c *galleryFileExistsCache) normalize(path string) string {
	return filepath.Clean(path)
}

// pruneLocked keeps the cache bounded: expired entries are removed when the map
// exceeds maxItems, and any remaining overflow is evicted arbitrarily (the
// value is simply re-derived from disk on the next listing). Caller holds mu.
func (c *galleryFileExistsCache) pruneLocked() {
	if len(c.items) <= c.maxItems {
		return
	}
	now := c.now()
	for k, e := range c.items {
		if now.Sub(e.at) >= c.ttl {
			delete(c.items, k)
		}
	}
	for k := range c.items {
		if len(c.items) <= c.maxItems {
			break
		}
		delete(c.items, k)
	}
}
