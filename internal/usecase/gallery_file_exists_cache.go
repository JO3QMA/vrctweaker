package usecase

import (
	"sync"
	"time"
)

// galleryFileStatCacheTTL bounds how long a cached file-existence result is
// reused by Gallery listing before the path is re-checked on disk. A short TTL
// keeps external deletions/restorations from staying stale for long while
// avoiding an os.Stat per row on every listing (ADR 0001).
const galleryFileStatCacheTTL = 30 * time.Second

type galleryFileStatEntry struct {
	exists bool
	at     time.Time
}

// galleryFileExistsCache memoizes per-path file-existence results used by
// Gallery listing. Entries expire after ttl; InvalidateAll / InvalidatePath
// drop entries early when the picture folder is synced or a file is ingested.
type galleryFileExistsCache struct {
	mu    sync.Mutex
	ttl   time.Duration
	now   func() time.Time
	items map[string]galleryFileStatEntry
}

func newGalleryFileExistsCache() *galleryFileExistsCache {
	return &galleryFileExistsCache{
		ttl:   galleryFileStatCacheTTL,
		now:   time.Now,
		items: make(map[string]galleryFileStatEntry),
	}
}

// get returns the cached existence for path when it is still fresh.
func (c *galleryFileExistsCache) get(path string) (exists bool, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.items[path]
	if !ok {
		return false, false
	}
	if c.now().Sub(e.at) >= c.ttl {
		delete(c.items, path)
		return false, false
	}
	return e.exists, true
}

// put records the existence of path as of now.
func (c *galleryFileExistsCache) put(path string, exists bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[path] = galleryFileStatEntry{exists: exists, at: c.now()}
}

// invalidateAll drops every cached entry. Called after a folder sync or ingest.
func (c *galleryFileExistsCache) invalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]galleryFileStatEntry)
}

// invalidatePath drops the entry for one path. Called when a file is ingested.
func (c *galleryFileExistsCache) invalidatePath(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, path)
}
