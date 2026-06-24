package identity

import "context"

// UserCacheRepository defines persistence for users_cache.
type UserCacheRepository interface {
	// List returns Listable friends (named API friends with status), not log-only stubs or unresolved pipeline presence.
	List(ctx context.Context) ([]*UserCache, error)
	GetByVRCUserID(ctx context.Context, vrcUserID string) (*UserCache, error)
	ListFavorites(ctx context.Context) ([]*UserCache, error)
	Save(ctx context.Context, u *UserCache) error
	SaveBatch(ctx context.Context, users []*UserCache) error
	Delete(ctx context.Context, vrcUserID string) error
	DeleteAll(ctx context.Context) (int64, error)
	// GetSelfBySessionFingerprint returns the cached self row for this token fingerprint, if any.
	GetSelfBySessionFingerprint(ctx context.Context, sessionFingerprint string) (*UserCache, error)
	// UpsertSelf replaces any existing self rows and inserts the given self profile.
	UpsertSelf(ctx context.Context, u *UserCache) error
	// DeleteSelfRows removes all user_kind=self rows (e.g. on logout).
	DeleteSelfRows(ctx context.Context) error
	// ListContactsNeedingProfileResolution returns contact rows with pipeline presence but no display name.
	ListContactsNeedingProfileResolution(ctx context.Context) ([]*UserCache, error)
}
