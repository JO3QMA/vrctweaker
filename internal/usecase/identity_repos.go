package usecase

import (
	"context"

	"vrchat-tweaker/internal/domain/identity"
)

// ponytail:#129 domain UserCacheRepository removed; boundary stays usecase-local.
type userCacheRepo interface {
	// List returns Listable friends (named API friends with status), not log-only stubs or unresolved pipeline presence.
	List(ctx context.Context) ([]*identity.UserCache, error)
	GetByVRCUserID(ctx context.Context, vrcUserID string) (*identity.UserCache, error)
	ListFavorites(ctx context.Context) ([]*identity.UserCache, error)
	Save(ctx context.Context, u *identity.UserCache) error
	SaveBatch(ctx context.Context, users []*identity.UserCache) error
	Delete(ctx context.Context, vrcUserID string) error
	DeleteAll(ctx context.Context) (int64, error)
	GetSelfBySessionFingerprint(ctx context.Context, sessionFingerprint string) (*identity.UserCache, error)
	UpsertSelf(ctx context.Context, u *identity.UserCache) error
	DeleteSelfRows(ctx context.Context) error
	ListContactsNeedingProfileResolution(ctx context.Context) ([]*identity.UserCache, error)
}
