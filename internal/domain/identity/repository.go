package identity

import "context"

// FriendCacheRepository defines persistence operations for friend cache.
type FriendCacheRepository interface {
	// List returns all cached friends.
	List(ctx context.Context) ([]*FriendCache, error)
	// GetByVRCUserID returns a cached friend by VRChat user ID.
	GetByVRCUserID(ctx context.Context, vrcUserID string) (*FriendCache, error)
	// ListFavorites returns cached friends marked as favorite.
	ListFavorites(ctx context.Context) ([]*FriendCache, error)
	// Save persists friend cache (upsert by VRCUserID).
	Save(ctx context.Context, f *FriendCache) error
	// SaveBatch persists multiple friends efficiently.
	SaveBatch(ctx context.Context, friends []*FriendCache) error
	// Delete removes a friend cache by VRChat user ID.
	Delete(ctx context.Context, vrcUserID string) error
	// DeleteAll removes all cached friends. Returns affected row count.
	DeleteAll(ctx context.Context) (int64, error)
}
