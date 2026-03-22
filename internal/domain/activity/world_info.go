package activity

import (
	"context"
	"time"
)

// WorldInfo is persisted world metadata from logs (keyed by wrld_* id).
type WorldInfo struct {
	WorldID       string
	DisplayName   string
	LastVisitedAt time.Time
}

// WorldInfoRepository persists world_info rows.
type WorldInfoRepository interface {
	UpsertVisit(ctx context.Context, worldID string, at time.Time) error
	UpsertDisplayName(ctx context.Context, worldID, displayName string, at time.Time) error
	GetByWorldID(ctx context.Context, worldID string) (*WorldInfo, error)
}
