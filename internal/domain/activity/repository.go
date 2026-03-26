package activity

import (
	"context"
	"time"
)

// PlaySessionRepository defines persistence operations for play sessions.
type PlaySessionRepository interface {
	// List returns play sessions within the given time range.
	List(ctx context.Context, from, to time.Time) ([]*PlaySession, error)
	// GetByID returns a play session by ID.
	GetByID(ctx context.Context, id string) (*PlaySession, error)
	// Save persists a play session.
	Save(ctx context.Context, s *PlaySession) error
	// FindLatestWithoutEndTime returns the most recent play session that has no end time.
	FindLatestWithoutEndTime(ctx context.Context) (*PlaySession, error)
	// Count returns the number of stored play sessions.
	Count(ctx context.Context) (int64, error)
}

// UserEncounterRepository defines persistence operations for user encounters.
type UserEncounterRepository interface {
	// List returns encounters with optional filters.
	List(ctx context.Context, filter *EncounterFilter) ([]*UserEncounter, error)
	// ListWithContext returns encounters with world and user cache fields.
	ListWithContext(ctx context.Context, filter *EncounterFilter) ([]*EncounterWithContext, error)
	// Save inserts a user encounter row (typically an open stay with LeftAt nil).
	Save(ctx context.Context, e *UserEncounter) error
	// CloseEncounterLeave sets left_at for the user's open stay (left_at IS NULL).
	CloseEncounterLeave(ctx context.Context, vrcUserID string, leftAt time.Time) (int64, error)
	// CloseOpenEncountersAt sets left_at for every row that is still open.
	CloseOpenEncountersAt(ctx context.Context, at time.Time) (int64, error)
	// DeleteOlderThan removes encounters older than the given time (for rotation).
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
	// DeleteAll removes all encounters. Returns affected row count.
	DeleteAll(ctx context.Context) (int64, error)
	// Count returns the number of stored encounters.
	Count(ctx context.Context) (int64, error)
	// BackfillMissingWorldContext sets world_id (and instance_id when empty) on rows with missing
	// world_id by propagating the previous row's non-empty context in joined_at ascending order.
	BackfillMissingWorldContext(ctx context.Context) (updated int64, err error)
}

// EncounterFilter provides optional filtering for List.
type EncounterFilter struct {
	VRCUserID   string
	DisplayName string
	InstanceID  string
	WorldID     string
	From        *time.Time
	To          *time.Time
}
