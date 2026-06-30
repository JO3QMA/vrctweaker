package usecase

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/identity"
)

const activityLogCheckpointKey = "activity_log_checkpoint"

// ActivityLogFileCheckpoint is per-output_log ingest progress.
type ActivityLogFileCheckpoint struct {
	ByteOffset     int64  `json:"byteOffset"`
	VRChatLineTime string `json:"vrChatLineTime,omitempty"`
}

// ActivityLogCheckpoint is persisted JSON in app_settings for incremental log ingest.
type ActivityLogCheckpoint struct {
	WatchPath string `json:"watchPath"`
	// Legacy single-file fields; migrated into Files on read.
	File           string                               `json:"file,omitempty"`
	ByteOffset     int64                                `json:"byteOffset,omitempty"`
	VRChatLineTime string                               `json:"vrChatLineTime,omitempty"`
	Files          map[string]ActivityLogFileCheckpoint `json:"files,omitempty"`
}

// NormalizeFiles migrates legacy single-file checkpoint fields into Files.
func (c *ActivityLogCheckpoint) NormalizeFiles() {
	if c == nil {
		return
	}
	if c.Files == nil {
		c.Files = make(map[string]ActivityLogFileCheckpoint)
	}
	if c.File != "" {
		if _, ok := c.Files[c.File]; !ok {
			c.Files[c.File] = ActivityLogFileCheckpoint{
				ByteOffset:     c.ByteOffset,
				VRChatLineTime: c.VRChatLineTime,
			}
		}
		c.File = ""
		c.ByteOffset = 0
		c.VRChatLineTime = ""
	}
}

// FileCheckpoint returns the checkpoint for a log file path.
func (c *ActivityLogCheckpoint) FileCheckpoint(path string) (ActivityLogFileCheckpoint, bool) {
	if c == nil {
		return ActivityLogFileCheckpoint{}, false
	}
	c.NormalizeFiles()
	fc, ok := c.Files[path]
	return fc, ok
}

// SetFileCheckpoint updates the checkpoint for one log file.
func (c *ActivityLogCheckpoint) SetFileCheckpoint(path string, fc ActivityLogFileCheckpoint) {
	c.NormalizeFiles()
	c.Files[path] = fc
}

func parseDateRange(fromISO, toISO string) (from, to time.Time, err error) {
	from, err = time.ParseInLocation("2006-01-02", fromISO, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err = time.ParseInLocation("2006-01-02", toISO, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	from = activity.StartOfLocalCalendarDay(from)
	// Inclusive toISO: range ends at the start of the next local calendar day (exclusive).
	to = activity.StartOfNextLocalCalendarDay(activity.StartOfLocalCalendarDay(to))
	return from, to, nil
}

// ActivityUseCase handles log parsing, play sessions, and user encounters.
type ActivityUseCase struct {
	playRepo      playSessionRepo
	encounterRepo userEncounterRepo
	settingsRepo  appSettingsRepo
	userCacheRepo userCacheRepo
	worldRepo     worldInfoRepo
}

// NewActivityUseCase creates a new ActivityUseCase.
func NewActivityUseCase(
	playRepo playSessionRepo,
	encounterRepo userEncounterRepo,
	settingsRepo appSettingsRepo,
	userCacheRepo userCacheRepo,
	worldRepo worldInfoRepo,
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

// ApplyCommand executes a fine-grained activity ingest command from SessionCorrelator.
func (uc *ActivityUseCase) ApplyCommand(ctx context.Context, logSource string, cmd any) error {
	if cmd == nil {
		return nil
	}
	switch c := cmd.(type) {
	case activity.EndPlaySessionCmd:
		return uc.EndPlaySession(ctx, logSource, c.At)
	case activity.StartPlaySessionCmd:
		return uc.StartPlaySession(ctx, logSource, c.InstanceID, c.At)
	case activity.CloseOpenEncountersAtCmd:
		return uc.CloseOpenEncountersAt(ctx, logSource, c.At)
	case activity.RecordEncounterJoinCmd:
		return uc.RecordEncounterAt(ctx, logSource, c.VRCUserID, c.DisplayName, activity.EncounterActionJoin, c.InstanceID, c.WorldID, c.At)
	case activity.RecordEncounterLeaveCmd:
		return uc.RecordEncounterAt(ctx, logSource, c.VRCUserID, c.DisplayName, activity.EncounterActionLeave, c.InstanceID, c.WorldID, c.At)
	case activity.UpsertWorldVisitCmd:
		return uc.UpsertWorldVisit(ctx, c.WorldID, c.At)
	case activity.UpsertWorldRoomNameCmd:
		return uc.UpsertWorldRoomName(ctx, c.WorldID, c.RoomName, c.At)
	default:
		return nil
	}
}

// RecordEncounter saves a join/leave event (uses current time).
func (uc *ActivityUseCase) RecordEncounter(ctx context.Context, vrcUserID, displayName, action, instanceID string) error {
	return uc.RecordEncounterAt(ctx, "", vrcUserID, displayName, action, instanceID, "", time.Now().UTC())
}

// RecordEncounterAt records a join as a new open stay or a leave as closing the user's open stay.
func (uc *ActivityUseCase) RecordEncounterAt(ctx context.Context, logSource, vrcUserID, displayName, action, instanceID, worldID string, at time.Time) error {
	wid := worldID
	if wid == "" && instanceID != "" {
		wid = activity.WorldIDFromInstanceKey(instanceID)
	}
	switch action {
	case activity.EncounterActionJoin:
		existing, err := uc.encounterRepo.FindByVRCUserIDAndJoinedAt(ctx, vrcUserID, at)
		if err != nil {
			return err
		}
		if existing != nil {
			patch := &activity.UserEncounter{
				ID:          existing.ID,
				VRCUserID:   existing.VRCUserID,
				DisplayName: displayName,
				InstanceID:  instanceID,
				WorldID:     wid,
				JoinedAt:    existing.JoinedAt,
				LeftAt:      existing.LeftAt,
			}
			if patchErr := uc.encounterRepo.UpdateEncounter(ctx, patch); patchErr != nil {
				return patchErr
			}
		} else {
			e := &activity.UserEncounter{
				ID:            uuid.New().String(),
				VRCUserID:     vrcUserID,
				DisplayName:   displayName,
				InstanceID:    instanceID,
				WorldID:       wid,
				LogSourcePath: logSource,
				JoinedAt:      at,
				LeftAt:        nil,
			}
			if err := uc.encounterRepo.Save(ctx, e); err != nil {
				return err
			}
		}
	case activity.EncounterActionLeave:
		if _, err := uc.encounterRepo.CloseEncounterLeave(ctx, vrcUserID, instanceID, at); err != nil {
			return err
		}
	default:
		return nil
	}
	if uc.userCacheRepo != nil && vrcUserID != "" {
		existing, err := uc.userCacheRepo.GetByVRCUserID(ctx, vrcUserID)
		if err != nil {
			return err
		}
		if existing == nil {
			existing = &identity.UserCache{VRCUserID: vrcUserID, UserKind: identity.UserKindContact}
		}
		existing.MergeFromLog(displayName, at)
		if err := uc.userCacheRepo.Save(ctx, existing); err != nil {
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
	c.NormalizeFiles()
	return &c, nil
}

// SetActivityLogFileCheckpoint updates one file entry in the checkpoint map.
func (uc *ActivityUseCase) SetActivityLogFileCheckpoint(ctx context.Context, watchPath, filePath string, byteOffset int64, vrChatLineTime string) error {
	cp, err := uc.GetActivityLogCheckpoint(ctx)
	if err != nil {
		return err
	}
	if cp == nil {
		cp = &ActivityLogCheckpoint{WatchPath: watchPath}
	}
	cp.WatchPath = watchPath
	cp.SetFileCheckpoint(filePath, ActivityLogFileCheckpoint{
		ByteOffset:     byteOffset,
		VRChatLineTime: vrChatLineTime,
	})
	return uc.SetActivityLogCheckpoint(ctx, cp)
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
func (uc *ActivityUseCase) StartPlaySession(ctx context.Context, logSource, instanceID string, startedAt time.Time) error {
	s := &activity.PlaySession{
		ID:            uuid.New().String(),
		StartTime:     startedAt,
		EndTime:       nil,
		InstanceID:    instanceID,
		LogSourcePath: logSource,
	}
	return uc.playRepo.Save(ctx, s)
}

// EndPlaySession closes the open session for the given log source.
func (uc *ActivityUseCase) EndPlaySession(ctx context.Context, logSource string, endedAt time.Time) error {
	open, err := uc.playRepo.FindOpenForLogSource(ctx, logSource)
	if err != nil || open == nil {
		return err
	}
	dur := int(endedAt.Sub(open.StartTime).Seconds())
	open.EndTime = &endedAt
	open.DurationSec = &dur
	return uc.playRepo.Save(ctx, open)
}

// CloseOpenEncountersAtLastLogLine sets left_at on any encounter row still open for logSource.
func (uc *ActivityUseCase) CloseOpenEncountersAtLastLogLine(ctx context.Context, logSource string, lastLine time.Time) error {
	if lastLine.IsZero() {
		return nil
	}
	_, err := uc.encounterRepo.CloseOpenEncountersAtForLogSource(ctx, logSource, lastLine)
	return err
}

// CloseOpenEncountersAtAll sets left_at on all open encounter rows (legacy NULL included).
func (uc *ActivityUseCase) CloseOpenEncountersAtAll(ctx context.Context, at time.Time) error {
	if at.IsZero() {
		return nil
	}
	_, err := uc.encounterRepo.CloseOpenEncountersAt(ctx, at)
	return err
}

// CloseOpenEncountersAt sets left_at on open encounter rows for one log source.
func (uc *ActivityUseCase) CloseOpenEncountersAt(ctx context.Context, logSource string, at time.Time) error {
	if at.IsZero() {
		return nil
	}
	_, err := uc.encounterRepo.CloseOpenEncountersAtForLogSource(ctx, logSource, at)
	return err
}

// FinalizeOpenActivityForLogSource closes open play sessions and encounters for one log source.
func (uc *ActivityUseCase) FinalizeOpenActivityForLogSource(ctx context.Context, logSource string, lastLine time.Time) error {
	if err := uc.CloseOpenPlaySessionAtLastLogLine(ctx, logSource, lastLine); err != nil {
		return err
	}
	return uc.CloseOpenEncountersAtLastLogLine(ctx, logSource, lastLine)
}

// FinalizeAllOpenActivity closes all open play sessions and encounters at lastLine (VRChat exit).
func (uc *ActivityUseCase) FinalizeAllOpenActivity(ctx context.Context, lastLine time.Time) error {
	if lastLine.IsZero() {
		return nil
	}
	opens, err := uc.playRepo.FindAllWithoutEndTime(ctx)
	if err != nil {
		return err
	}
	for _, open := range opens {
		if err := uc.closePlaySessionAt(ctx, open, lastLine); err != nil {
			return err
		}
	}
	_, err = uc.encounterRepo.CloseOpenEncountersAt(ctx, lastLine)
	return err
}

// CloseOpenPlaySessionAtLastLogLine closes the open play session for logSource at lastLine.
func (uc *ActivityUseCase) CloseOpenPlaySessionAtLastLogLine(ctx context.Context, logSource string, lastLine time.Time) error {
	if lastLine.IsZero() {
		return nil
	}
	open, err := uc.playRepo.FindOpenForLogSource(ctx, logSource)
	if err != nil || open == nil {
		return err
	}
	return uc.closePlaySessionAt(ctx, open, lastLine)
}

func (uc *ActivityUseCase) closePlaySessionAt(ctx context.Context, open *activity.PlaySession, lastLine time.Time) error {
	if open == nil || lastLine.IsZero() {
		return nil
	}
	if lastLine.Before(open.StartTime) {
		return nil
	}
	if activity.SameLocalCalendarDay(open.StartTime, lastLine) {
		dur := int(lastLine.Sub(open.StartTime).Seconds())
		open.EndTime = &lastLine
		open.DurationSec = &dur
		return uc.playRepo.Save(ctx, open)
	}
	cur := open.StartTime
	id := open.ID
	logSource := open.LogSourcePath
	instanceID := open.InstanceID
	for {
		if activity.SameLocalCalendarDay(cur, lastLine) {
			dur := int(lastLine.Sub(cur).Seconds())
			s := &activity.PlaySession{
				ID:            id,
				StartTime:     cur,
				EndTime:       &lastLine,
				DurationSec:   &dur,
				InstanceID:    instanceID,
				LogSourcePath: logSource,
			}
			return uc.playRepo.Save(ctx, s)
		}
		segEnd := activity.EndOfLocalCalendarDay(cur)
		dur := int(segEnd.Sub(cur).Seconds())
		s := &activity.PlaySession{
			ID:            id,
			StartTime:     cur,
			EndTime:       &segEnd,
			DurationSec:   &dur,
			InstanceID:    instanceID,
			LogSourcePath: logSource,
		}
		if err := uc.playRepo.Save(ctx, s); err != nil {
			return err
		}
		cur = activity.StartOfNextLocalCalendarDay(cur)
		id = uuid.New().String()
	}
}

// GetActivityStats returns aggregated play stats for the date range [fromISO, toISO].
// fromISO, toISO are date strings in YYYY-MM-DD format (local calendar).
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
	if opens, err := uc.playRepo.FindAllWithoutEndTime(ctx); err != nil {
		return nil, err
	} else {
		seen := make(map[string]bool, len(sessions))
		for _, s := range sessions {
			seen[s.ID] = true
		}
		for _, open := range opens {
			if !seen[open.ID] {
				sessions = append(sessions, open)
			}
		}
	}
	lastObserved := uc.lastObservedLogTime(ctx)
	daily, topWorlds := activity.AggregatePlaySessions(sessions, from, to, lastObserved)
	return &activity.ActivityStats{
		DailyPlaySeconds: daily,
		TopWorlds:        topWorlds,
	}, nil
}

func (uc *ActivityUseCase) lastObservedLogTime(ctx context.Context) *time.Time {
	cp, err := uc.GetActivityLogCheckpoint(ctx)
	if err != nil || cp == nil {
		return nil
	}
	cp.NormalizeFiles()
	var max time.Time
	for _, fc := range cp.Files {
		if fc.VRChatLineTime == "" {
			continue
		}
		t, err := time.Parse(time.RFC3339, fc.VRChatLineTime)
		if err != nil || t.IsZero() {
			continue
		}
		if t.After(max) {
			max = t
		}
	}
	if max.IsZero() {
		return nil
	}
	return &max
}

func (uc *ActivityUseCase) activityRetentionCutoff(ctx context.Context) (time.Time, error) {
	daysStr, err := uc.settingsRepo.Get(ctx, "log_retention_days")
	if err != nil {
		return time.Time{}, err
	}
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}
	return time.Now().UTC().AddDate(0, 0, -days), nil
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

// DeduplicateEncounters merges duplicate encounter rows and fixes invalid leave times.
func (uc *ActivityUseCase) DeduplicateEncounters(ctx context.Context) (int64, error) {
	return uc.encounterRepo.DeduplicateEncounters(ctx)
}

// RotateEncounters deletes encounters and play sessions older than Activity retention days.
func (uc *ActivityUseCase) RotateEncounters(ctx context.Context) (int64, error) {
	before, err := uc.activityRetentionCutoff(ctx)
	if err != nil {
		return 0, err
	}
	encDeleted, err := uc.encounterRepo.DeleteOlderThan(ctx, before)
	if err != nil {
		return 0, err
	}
	playDeleted, err := uc.playRepo.DeleteOlderThan(ctx, before)
	if err != nil {
		return 0, err
	}
	return encDeleted + playDeleted, nil
}
