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
	// Save persists a user encounter.
	Save(ctx context.Context, e *UserEncounter) error
	// DeleteOlderThan removes encounters older than the given time (for rotation).
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
	// DeleteAll removes all encounters. Returns affected row count.
	DeleteAll(ctx context.Context) (int64, error)
	// Count returns the number of stored encounters.
	Count(ctx context.Context) (int64, error)
}

// EncounterFilter provides optional filtering for List.
type EncounterFilter struct {
	VRCUserID   string
	DisplayName string
	InstanceID  string
	From        *time.Time
	To          *time.Time
}
