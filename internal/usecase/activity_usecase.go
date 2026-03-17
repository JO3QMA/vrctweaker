package usecase

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/settings"
)

func parseDateRange(fromISO, toISO string) (from, to time.Time, err error) {
	from, err = time.ParseInLocation("2006-01-02", fromISO, time.UTC)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err = time.ParseInLocation("2006-01-02", toISO, time.UTC)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, time.UTC)
	return from, to, nil
}

// ActivityUseCase handles log parsing, play sessions, and user encounters.
type ActivityUseCase struct {
	playRepo      activity.PlaySessionRepository
	encounterRepo activity.UserEncounterRepository
	settingsRepo  settings.AppSettingsRepository
}

// NewActivityUseCase creates a new ActivityUseCase.
func NewActivityUseCase(
	playRepo activity.PlaySessionRepository,
	encounterRepo activity.UserEncounterRepository,
	settingsRepo settings.AppSettingsRepository,
) *ActivityUseCase {
	return &ActivityUseCase{
		playRepo:      playRepo,
		encounterRepo: encounterRepo,
		settingsRepo:  settingsRepo,
	}
}

// ListEncounters returns user encounters with optional filter.
func (uc *ActivityUseCase) ListEncounters(ctx context.Context, filter *activity.EncounterFilter) ([]*activity.UserEncounter, error) {
	return uc.encounterRepo.List(ctx, filter)
}

// RecordEncounter saves a join/leave event (uses current time).
func (uc *ActivityUseCase) RecordEncounter(ctx context.Context, vrcUserID, displayName, action, instanceID string) error {
	return uc.RecordEncounterAt(ctx, vrcUserID, displayName, action, instanceID, time.Now().UTC())
}

// RecordEncounterAt saves a join/leave event with explicit timestamp.
func (uc *ActivityUseCase) RecordEncounterAt(ctx context.Context, vrcUserID, displayName, action, instanceID string, at time.Time) error {
	e := &activity.UserEncounter{
		ID:            uuid.New().String(),
		VRCUserID:     vrcUserID,
		DisplayName:   displayName,
		Action:        action,
		InstanceID:    instanceID,
		EncounteredAt: at,
	}
	return uc.encounterRepo.Save(ctx, e)
}

// ListPlaySessions returns play sessions in the time range.
func (uc *ActivityUseCase) ListPlaySessions(ctx context.Context, from, to time.Time) ([]*activity.PlaySession, error) {
	return uc.playRepo.List(ctx, from, to)
}

// SavePlaySession persists a play session.
func (uc *ActivityUseCase) SavePlaySession(ctx context.Context, s *activity.PlaySession) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return uc.playRepo.Save(ctx, s)
}

// StartPlaySession creates a new play session with the given start time.
func (uc *ActivityUseCase) StartPlaySession(ctx context.Context, instanceID string, startedAt time.Time) error {
	_ = instanceID // reserved for future schema extension
	s := &activity.PlaySession{
		ID:        uuid.New().String(),
		StartTime: startedAt,
		EndTime:   nil,
	}
	return uc.playRepo.Save(ctx, s)
}

// EndPlaySession closes the most recent open session with end time and duration.
func (uc *ActivityUseCase) EndPlaySession(ctx context.Context, endedAt time.Time) error {
	open, err := uc.playRepo.FindLatestWithoutEndTime(ctx)
	if err != nil || open == nil {
		return err
	}
	dur := int(endedAt.Sub(open.StartTime).Seconds())
	open.EndTime = &endedAt
	open.DurationSec = &dur
	return uc.playRepo.Save(ctx, open)
}

// GetActivityStats returns aggregated play stats for the date range [fromISO, toISO].
// fromISO, toISO are date strings in YYYY-MM-DD format.
func (uc *ActivityUseCase) GetActivityStats(ctx context.Context, fromISO, toISO string) (*activity.ActivityStats, error) {
	from, to, err := parseDateRange(fromISO, toISO)
	if err != nil {
		return nil, err
	}
	// Expand from by 24h to catch sessions starting before range but ending within it
	sessions, err := uc.playRepo.List(ctx, from.Add(-24*time.Hour), to)
	if err != nil {
		return nil, err
	}
	daily, topWorlds := activity.AggregatePlaySessions(sessions, from, to)
	return &activity.ActivityStats{
		DailyPlaySeconds: daily,
		TopWorlds:        topWorlds,
	}, nil
}

// RotateEncounters deletes encounters older than retention days.
func (uc *ActivityUseCase) RotateEncounters(ctx context.Context) (int64, error) {
	daysStr, err := uc.settingsRepo.Get(ctx, "log_retention_days")
	if err != nil {
		return 0, err
	}
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}
	before := time.Now().UTC().AddDate(0, 0, -days)
	return uc.encounterRepo.DeleteOlderThan(ctx, before)
}
