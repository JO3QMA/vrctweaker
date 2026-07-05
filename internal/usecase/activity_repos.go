package usecase

import (
	"context"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

// ponytail:#129 domain activity repository interfaces removed; boundaries stay usecase-local.

type playSessionRepo interface {
	List(ctx context.Context, from, to time.Time) ([]*activity.PlaySession, error)
	GetByID(ctx context.Context, id string) (*activity.PlaySession, error)
	Save(ctx context.Context, s *activity.PlaySession) error
	FindLatestWithoutEndTime(ctx context.Context) (*activity.PlaySession, error)
	FindOpenForLogSource(ctx context.Context, logSource string) (*activity.PlaySession, error)
	FindAllWithoutEndTime(ctx context.Context) ([]*activity.PlaySession, error)
	Count(ctx context.Context) (int64, error)
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}

type userEncounterRepo interface {
	List(ctx context.Context, filter *activity.EncounterFilter) ([]*activity.UserEncounter, error)
	ListWithContext(ctx context.Context, filter *activity.EncounterFilter) ([]*activity.EncounterWithContext, error)
	Save(ctx context.Context, e *activity.UserEncounter) error
	FindByVRCUserIDAndJoinedAt(ctx context.Context, vrcUserID string, joinedAt time.Time) (*activity.UserEncounter, error)
	UpdateEncounter(ctx context.Context, e *activity.UserEncounter) error
	CloseEncounterLeave(ctx context.Context, vrcUserID, instanceID string, leftAt time.Time) (int64, error)
	CloseOpenEncountersAt(ctx context.Context, at time.Time) (int64, error)
	CloseOpenEncountersAtForLogSource(ctx context.Context, logSource string, at time.Time) (int64, error)
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
	DeleteAll(ctx context.Context) (int64, error)
	Count(ctx context.Context) (int64, error)
	BackfillMissingWorldContext(ctx context.Context) (int64, error)
	DeduplicateEncounters(ctx context.Context) (int64, error)
}

type worldInfoRepo interface {
	UpsertVisit(ctx context.Context, worldID string, at time.Time) error
	UpsertDisplayName(ctx context.Context, worldID, displayName string, at time.Time) error
	GetByWorldID(ctx context.Context, worldID string) (*activity.WorldInfo, error)
}
