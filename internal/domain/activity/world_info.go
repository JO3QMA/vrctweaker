package activity

import "time"

// WorldInfo is persisted world metadata from logs (keyed by wrld_* id).
type WorldInfo struct {
	WorldID       string
	DisplayName   string
	LastVisitedAt time.Time
}
