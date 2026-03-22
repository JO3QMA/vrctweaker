package usecase

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/domain/settings"
)

const activityLogCheckpointKey = "activity_log_checkpoint"

// ActivityLogCheckpoint is persisted JSON in app_settings for incremental log ingest.
type ActivityLogCheckpoint struct {
	WatchPath      string `json:"watchPath"`
	File           string `json:"file"`
	ByteOffset     int64  `json:"byteOffset"`
	VRChatLineTime string `json:"vrChatLineTime,omitempty"`
}

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
	userCacheRepo identity.UserCacheRepository
	worldRepo     activity.WorldInfoRepository
}

// NewActivityUseCase creates a new ActivityUseCase.
func NewActivityUseCase(
	playRepo activity.PlaySessionRepository,
	encounterRepo activity.UserEncounterRepository,
	settingsRepo settings.AppSettingsRepository,
	userCacheRepo identity.UserCacheRepository,
	worldRepo activity.WorldInfoRepository,
) *ActivityUseCase {
	return &ActivityUseCase{
		playRepo:      playRepo,
		encounterRepo: encounterRepo,
		settingsRepo:  settingsRepo,
		userCacheRepo: userCacheRepo,
		worldRepo:     worldRepo,
	}
}

// ListEncounters returns user encounters with optional filter.
func (uc *ActivityUseCase) ListEncounters(ctx context.Context, filter *activity.EncounterFilter) ([]*activity.UserEncounter, error) {
	return uc.encounterRepo.List(ctx, filter)
}

// ListEncountersWithContext returns encounters joined with user/world cache for the UI.
func (uc *ActivityUseCase) ListEncountersWithContext(ctx context.Context, filter *activity.EncounterFilter) ([]*activity.EncounterWithContext, error) {
	rows, err := uc.encounterRepo.ListWithContext(ctx, filter)
	if err != nil {
		return nil, err
	}
	if uc.worldRepo == nil {
		return rows, nil
	}
	for _, row := range rows {
		enc := row.Encounter
		if row.WorldDisplayName != "" {
			continue
		}
		wid := enc.WorldID
		if wid == "" {
			wid = activity.WorldIDFromInstanceKey(enc.InstanceID)
		}
		if wid == "" {
			continue
		}
		wi, err := uc.worldRepo.GetByWorldID(ctx, wid)
		if err != nil || wi == nil || wi.DisplayName == "" {
			continue
		}
		row.WorldDisplayName = wi.DisplayName
	}
	return rows, nil
}

// RecordEncounter saves a join/leave event (uses current time).
func (uc *ActivityUseCase) RecordEncounter(ctx context.Context, vrcUserID, displayName, action, instanceID string) error {
	return uc.RecordEncounterAt(ctx, vrcUserID, displayName, action, instanceID, "", time.Now().UTC())
}

// RecordEncounterAt saves a join/leave event with explicit timestamp and optional world id.
func (uc *ActivityUseCase) RecordEncounterAt(ctx context.Context, vrcUserID, displayName, action, instanceID, worldID string, at time.Time) error {
	wid := worldID
	if wid == "" && instanceID != "" {
		wid = activity.WorldIDFromInstanceKey(instanceID)
	}
	e := &activity.UserEncounter{
		ID:            uuid.New().String(),
		VRCUserID:     vrcUserID,
		DisplayName:   displayName,
		Action:        action,
		InstanceID:    instanceID,
		WorldID:       wid,
		EncounteredAt: at,
	}
	if err := uc.encounterRepo.Save(ctx, e); err != nil {
		return err
	}
	if uc.userCacheRepo != nil && vrcUserID != "" {
		if err := uc.userCacheRepo.UpsertFromLog(ctx, vrcUserID, displayName, at); err != nil {
			return err
		}
	}
	return nil
}

// GetActivityLogCheckpoint loads the last processed log position.
func (uc *ActivityUseCase) GetActivityLogCheckpoint(ctx context.Context) (*ActivityLogCheckpoint, error) {
	raw, err := uc.settingsRepo.Get(ctx, activityLogCheckpointKey)
	if err != nil || raw == "" {
		return nil, nil
	}
	var c ActivityLogCheckpoint
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// SetActivityLogCheckpoint persists the last processed log position.
func (uc *ActivityUseCase) SetActivityLogCheckpoint(ctx context.Context, c *ActivityLogCheckpoint) error {
	if c == nil {
		return uc.settingsRepo.Set(ctx, activityLogCheckpointKey, "")
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return uc.settingsRepo.Set(ctx, activityLogCheckpointKey, string(b))
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

// CloseOpenPlaySessionAtLastLogLine closes the latest open play session using the timestamp of the
// last processed log line. Same local calendar day: single segment [start, lastLine]. If the session
// spans local midnights, it is split into multiple closed rows at each local 23:59:59.999999999 with
// continuation rows starting at the next local midnight (VRChat may keep writing to the previous
// day's output_log file after date change).
func (uc *ActivityUseCase) CloseOpenPlaySessionAtLastLogLine(ctx context.Context, lastLine time.Time) error {
	if lastLine.IsZero() {
		return nil
	}
	open, err := uc.playRepo.FindLatestWithoutEndTime(ctx)
	if err != nil || open == nil {
		return err
	}
	if lastLine.Before(open.StartTime) {
		return nil
	}
	if activity.SameLocalCalendarDay(open.StartTime, lastLine) {
		return uc.EndPlaySession(ctx, lastLine)
	}
	cur := open.StartTime
	id := open.ID
	for {
		if activity.SameLocalCalendarDay(cur, lastLine) {
			dur := int(lastLine.Sub(cur).Seconds())
			s := &activity.PlaySession{
				ID:          id,
				StartTime:   cur,
				EndTime:     &lastLine,
				DurationSec: &dur,
			}
			return uc.playRepo.Save(ctx, s)
		}
		segEnd := activity.EndOfLocalCalendarDay(cur)
		dur := int(segEnd.Sub(cur).Seconds())
		s := &activity.PlaySession{
			ID:          id,
			StartTime:   cur,
			EndTime:     &segEnd,
			DurationSec: &dur,
		}
		if err := uc.playRepo.Save(ctx, s); err != nil {
			return err
		}
		cur = activity.StartOfNextLocalCalendarDay(cur)
		id = uuid.New().String()
	}
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

// IsActivityDatastoreEmpty reports whether both play sessions and encounters are absent.
func (uc *ActivityUseCase) IsActivityDatastoreEmpty(ctx context.Context) (bool, error) {
	pc, err := uc.playRepo.Count(ctx)
	if err != nil {
		return false, err
	}
	ec, err := uc.encounterRepo.Count(ctx)
	if err != nil {
		return false, err
	}
	return pc == 0 && ec == 0, nil
}

// UpsertWorldVisit records a world visit from log lines (Destination set).
func (uc *ActivityUseCase) UpsertWorldVisit(ctx context.Context, worldID string, at time.Time) error {
	if uc.worldRepo == nil || worldID == "" {
		return nil
	}
	return uc.worldRepo.UpsertVisit(ctx, worldID, at)
}

// UpsertWorldRoomName sets display name from Entering Room lines.
func (uc *ActivityUseCase) UpsertWorldRoomName(ctx context.Context, worldID, roomName string, at time.Time) error {
	if uc.worldRepo == nil || worldID == "" || roomName == "" {
		return nil
	}
	return uc.worldRepo.UpsertDisplayName(ctx, worldID, roomName, at)
}

// BackfillEncounterWorldContext fills missing world_id (and instance_id when empty) on stored
// encounters by propagating the previous row's context in time order.
func (uc *ActivityUseCase) BackfillEncounterWorldContext(ctx context.Context) (int64, error) {
	return uc.encounterRepo.BackfillMissingWorldContext(ctx)
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
