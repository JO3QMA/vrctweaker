package usecase

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/identity"
)

type fakePlaySessionRepo struct {
	sessions []*activity.PlaySession
}

func (f *fakePlaySessionRepo) List(_ context.Context, _, _ time.Time) ([]*activity.PlaySession, error) {
	out := make([]*activity.PlaySession, len(f.sessions))
	copy(out, f.sessions)
	return out, nil
}

func (f *fakePlaySessionRepo) GetByID(context.Context, string) (*activity.PlaySession, error) {
	return nil, nil
}

func (f *fakePlaySessionRepo) Save(_ context.Context, s *activity.PlaySession) error {
	for i, ex := range f.sessions {
		if ex.ID == s.ID {
			f.sessions[i] = s
			return nil
		}
	}
	f.sessions = append(f.sessions, s)
	return nil
}

func (f *fakePlaySessionRepo) FindLatestWithoutEndTime(context.Context) (*activity.PlaySession, error) {
	for i := len(f.sessions) - 1; i >= 0; i-- {
		if f.sessions[i].EndTime == nil {
			return f.sessions[i], nil
		}
	}
	return nil, nil
}

func (f *fakePlaySessionRepo) FindOpenForLogSource(_ context.Context, logSource string) (*activity.PlaySession, error) {
	for i := len(f.sessions) - 1; i >= 0; i-- {
		if f.sessions[i].EndTime == nil && f.sessions[i].LogSourcePath == logSource {
			return f.sessions[i], nil
		}
	}
	return nil, nil
}

func (f *fakePlaySessionRepo) FindAllWithoutEndTime(_ context.Context) ([]*activity.PlaySession, error) {
	var out []*activity.PlaySession
	for _, s := range f.sessions {
		if s.EndTime == nil {
			cp := *s
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (f *fakePlaySessionRepo) Count(context.Context) (int64, error) {
	return int64(len(f.sessions)), nil
}

func (f *fakePlaySessionRepo) DeleteOlderThan(_ context.Context, before time.Time) (int64, error) {
	var kept []*activity.PlaySession
	var deleted int64
	for _, s := range f.sessions {
		if s.StartTime.Before(before) {
			deleted++
			continue
		}
		kept = append(kept, s)
	}
	f.sessions = kept
	return deleted, nil
}

type stubEncounterRepo struct{}

func (stubEncounterRepo) List(context.Context, *activity.EncounterFilter) ([]*activity.UserEncounter, error) {
	return nil, nil
}

func (stubEncounterRepo) ListWithContext(context.Context, *activity.EncounterFilter) ([]*activity.EncounterWithContext, error) {
	return nil, nil
}

func (stubEncounterRepo) Save(context.Context, *activity.UserEncounter) error { return nil }

func (stubEncounterRepo) FindByVRCUserIDAndJoinedAt(context.Context, string, time.Time) (*activity.UserEncounter, error) {
	return nil, nil
}

func (stubEncounterRepo) UpdateEncounter(context.Context, *activity.UserEncounter) error { return nil }

func (stubEncounterRepo) CloseEncounterLeave(context.Context, string, string, time.Time) (int64, error) {
	return 0, nil
}

func (stubEncounterRepo) CloseOpenEncountersAt(context.Context, time.Time) (int64, error) {
	return 0, nil
}

func (stubEncounterRepo) CloseOpenEncountersAtForLogSource(context.Context, string, time.Time) (int64, error) {
	return 0, nil
}

func (stubEncounterRepo) DeleteOlderThan(context.Context, time.Time) (int64, error) { return 0, nil }

func (stubEncounterRepo) DeleteAll(context.Context) (int64, error) { return 0, nil }

func (stubEncounterRepo) Count(context.Context) (int64, error) { return 0, nil }

func (stubEncounterRepo) BackfillMissingWorldContext(context.Context) (int64, error) { return 0, nil }

func (stubEncounterRepo) DeduplicateEncounters(context.Context) (int64, error) { return 0, nil }

type recordingEncounterRepo struct {
	stubEncounterRepo
	closeOpenAts      []time.Time
	closeForLogSource []struct {
		logSource string
		at        time.Time
	}
}

func (r *recordingEncounterRepo) CloseOpenEncountersAt(_ context.Context, at time.Time) (int64, error) {
	r.closeOpenAts = append(r.closeOpenAts, at)
	return 0, nil
}

func (r *recordingEncounterRepo) CloseOpenEncountersAtForLogSource(_ context.Context, logSource string, at time.Time) (int64, error) {
	r.closeForLogSource = append(r.closeForLogSource, struct {
		logSource string
		at        time.Time
	}{logSource, at})
	return 0, nil
}

func TestActivityUseCase_CloseOpenPlaySessionAtLastLogLine_sameDay(t *testing.T) {
	ctx := context.Background()
	const logSource = "/logs/output_log.txt"
	repo := &fakePlaySessionRepo{}
	uc := NewActivityUseCase(repo, stubEncounterRepo{}, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)

	start := time.Date(2024, 3, 18, 10, 0, 0, 0, time.Local)
	_ = repo.Save(ctx, &activity.PlaySession{
		ID:            "s1",
		StartTime:     start,
		EndTime:       nil,
		LogSourcePath: logSource,
	})

	lastLine := time.Date(2024, 3, 18, 20, 0, 0, 0, time.Local)
	if err := uc.CloseOpenPlaySessionAtLastLogLine(ctx, logSource, lastLine); err != nil {
		t.Fatalf("CloseOpenPlaySessionAtLastLogLine: %v", err)
	}
	open, err := repo.FindOpenForLogSource(ctx, logSource)
	if err != nil {
		t.Fatalf("FindLatestWithoutEndTime: %v", err)
	}
	if open != nil {
		t.Fatalf("expected session closed, still open: %+v", open)
	}
	if len(repo.sessions) != 1 {
		t.Fatalf("sessions len = %d, want 1", len(repo.sessions))
	}
	if !repo.sessions[0].EndTime.Equal(lastLine) {
		t.Fatalf("EndTime = %v, want %v", repo.sessions[0].EndTime, lastLine)
	}
}

func TestActivityUseCase_CloseOpenPlaySessionAtLastLogLine_splitsTwoLocalDays(t *testing.T) {
	ctx := context.Background()
	const logSource = "/logs/output_log.txt"
	repo := &fakePlaySessionRepo{}
	uc := NewActivityUseCase(repo, stubEncounterRepo{}, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)

	start := time.Date(2024, 3, 18, 23, 0, 0, 0, time.Local)
	_ = repo.Save(ctx, &activity.PlaySession{
		ID:            "s1",
		StartTime:     start,
		EndTime:       nil,
		LogSourcePath: logSource,
	})

	lastLine := time.Date(2024, 3, 19, 1, 0, 0, 0, time.Local)
	if err := uc.CloseOpenPlaySessionAtLastLogLine(ctx, logSource, lastLine); err != nil {
		t.Fatalf("CloseOpenPlaySessionAtLastLogLine: %v", err)
	}
	open, err := repo.FindOpenForLogSource(ctx, logSource)
	if err != nil {
		t.Fatalf("FindLatestWithoutEndTime: %v", err)
	}
	if open != nil {
		t.Fatal("expected all segments closed")
	}
	if len(repo.sessions) != 2 {
		t.Fatalf("sessions len = %d, want 2", len(repo.sessions))
	}
	wantEnd0 := activity.EndOfLocalCalendarDay(start)
	if !repo.sessions[0].EndTime.Equal(wantEnd0) {
		t.Fatalf("segment0 EndTime = %v, want %v", repo.sessions[0].EndTime, wantEnd0)
	}
	wantStart1 := activity.StartOfNextLocalCalendarDay(start)
	if !repo.sessions[1].StartTime.Equal(wantStart1) {
		t.Fatalf("segment1 StartTime = %v, want %v", repo.sessions[1].StartTime, wantStart1)
	}
	if !repo.sessions[1].EndTime.Equal(lastLine) {
		t.Fatalf("segment1 EndTime = %v, want %v", repo.sessions[1].EndTime, lastLine)
	}
}

func TestActivityUseCase_CloseOpenPlaySessionAtLastLogLine_splitsThreeLocalDays(t *testing.T) {
	ctx := context.Background()
	const logSource = "/logs/output_log.txt"
	repo := &fakePlaySessionRepo{}
	uc := NewActivityUseCase(repo, stubEncounterRepo{}, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)

	start := time.Date(2024, 3, 18, 20, 0, 0, 0, time.Local)
	_ = repo.Save(ctx, &activity.PlaySession{
		ID:            "s1",
		StartTime:     start,
		EndTime:       nil,
		LogSourcePath: logSource,
	})

	lastLine := time.Date(2024, 3, 20, 3, 0, 0, 0, time.Local)
	if err := uc.CloseOpenPlaySessionAtLastLogLine(ctx, logSource, lastLine); err != nil {
		t.Fatalf("CloseOpenPlaySessionAtLastLogLine: %v", err)
	}
	open3, err3 := repo.FindOpenForLogSource(ctx, logSource)
	if err3 != nil {
		t.Fatal(err3)
	}
	if open3 != nil {
		t.Fatal("expected all segments closed")
	}
	if len(repo.sessions) != 3 {
		t.Fatalf("sessions len = %d, want 3: %#v", len(repo.sessions), repo.sessions)
	}
	if !repo.sessions[0].EndTime.Equal(activity.EndOfLocalCalendarDay(start)) {
		t.Fatalf("segment0 end = %v", repo.sessions[0].EndTime)
	}
	mid := activity.StartOfNextLocalCalendarDay(start)
	if !repo.sessions[1].StartTime.Equal(mid) {
		t.Fatalf("segment1 start = %v, want %v", repo.sessions[1].StartTime, mid)
	}
	if !repo.sessions[1].EndTime.Equal(activity.EndOfLocalCalendarDay(mid)) {
		t.Fatalf("segment1 end = %v", repo.sessions[1].EndTime)
	}
	if !repo.sessions[2].EndTime.Equal(lastLine) {
		t.Fatalf("segment2 end = %v, want %v", repo.sessions[2].EndTime, lastLine)
	}
}

type memEncounterRepo struct {
	encounters []*activity.UserEncounter
	contexts   []*activity.EncounterWithContext
	backfillN  int64
}

func (m *memEncounterRepo) List(_ context.Context, _ *activity.EncounterFilter) ([]*activity.UserEncounter, error) {
	out := make([]*activity.UserEncounter, len(m.encounters))
	copy(out, m.encounters)
	return out, nil
}

func (m *memEncounterRepo) ListWithContext(_ context.Context, _ *activity.EncounterFilter) ([]*activity.EncounterWithContext, error) {
	if m.contexts != nil {
		return m.contexts, nil
	}
	return nil, nil
}

func (m *memEncounterRepo) Save(_ context.Context, e *activity.UserEncounter) error {
	cp := *e
	m.encounters = append(m.encounters, &cp)
	return nil
}

func (m *memEncounterRepo) FindByVRCUserIDAndJoinedAt(_ context.Context, vrcUserID string, joinedAt time.Time) (*activity.UserEncounter, error) {
	for _, e := range m.encounters {
		if e.VRCUserID == vrcUserID && e.JoinedAt.Equal(joinedAt) {
			cp := *e
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *memEncounterRepo) UpdateEncounter(_ context.Context, e *activity.UserEncounter) error {
	for _, ex := range m.encounters {
		if ex.ID != e.ID {
			continue
		}
		ex.DisplayName = e.DisplayName
		if ex.InstanceID == "" && e.InstanceID != "" {
			ex.InstanceID = e.InstanceID
		}
		if ex.WorldID == "" && e.WorldID != "" {
			ex.WorldID = e.WorldID
		}
		return nil
	}
	return nil
}

func (m *memEncounterRepo) CloseEncounterLeave(_ context.Context, vrcUserID, instanceID string, leftAt time.Time) (int64, error) {
	var target *activity.UserEncounter
	for _, e := range m.encounters {
		if e.VRCUserID != vrcUserID || e.LeftAt != nil || e.JoinedAt.After(leftAt) {
			continue
		}
		if instanceID != "" && e.InstanceID != "" && e.InstanceID != instanceID {
			continue
		}
		if target == nil || e.JoinedAt.After(target.JoinedAt) {
			target = e
		}
	}
	if target == nil {
		return 0, nil
	}
	t := leftAt
	target.LeftAt = &t
	return 1, nil
}

func (m *memEncounterRepo) CloseOpenEncountersAt(_ context.Context, at time.Time) (int64, error) {
	var n int64
	for _, e := range m.encounters {
		if e.LeftAt == nil && !e.JoinedAt.After(at) {
			t := at
			e.LeftAt = &t
			n++
		}
	}
	return n, nil
}

func (m *memEncounterRepo) CloseOpenEncountersAtForLogSource(_ context.Context, logSource string, at time.Time) (int64, error) {
	var n int64
	for _, e := range m.encounters {
		if e.LeftAt == nil && e.LogSourcePath == logSource && !e.JoinedAt.After(at) {
			t := at
			e.LeftAt = &t
			n++
		}
	}
	return n, nil
}

func (m *memEncounterRepo) DeleteOlderThan(_ context.Context, before time.Time) (int64, error) {
	var kept []*activity.UserEncounter
	var n int64
	for _, e := range m.encounters {
		if e.JoinedAt.Before(before) {
			n++
		} else {
			kept = append(kept, e)
		}
	}
	m.encounters = kept
	return n, nil
}

func (m *memEncounterRepo) DeleteAll(_ context.Context) (int64, error) {
	n := int64(len(m.encounters))
	m.encounters = nil
	return n, nil
}

func (m *memEncounterRepo) Count(_ context.Context) (int64, error) {
	return int64(len(m.encounters)), nil
}

func (m *memEncounterRepo) BackfillMissingWorldContext(_ context.Context) (int64, error) {
	return m.backfillN, nil
}

func (m *memEncounterRepo) DeduplicateEncounters(context.Context) (int64, error) {
	return 0, nil
}

type activityUserCacheRepo struct {
	byID map[string]*identity.UserCache
}

func (r *activityUserCacheRepo) List(_ context.Context) ([]*identity.UserCache, error) {
	return nil, nil
}
func (r *activityUserCacheRepo) GetByVRCUserID(_ context.Context, id string) (*identity.UserCache, error) {
	if r.byID == nil {
		return nil, nil
	}
	return r.byID[id], nil
}
func (r *activityUserCacheRepo) ListFavorites(_ context.Context) ([]*identity.UserCache, error) {
	return nil, nil
}
func (r *activityUserCacheRepo) Save(_ context.Context, u *identity.UserCache) error {
	if r.byID == nil {
		r.byID = make(map[string]*identity.UserCache)
	}
	cp := *u
	r.byID[u.VRCUserID] = &cp
	return nil
}
func (r *activityUserCacheRepo) SaveBatch(_ context.Context, _ []*identity.UserCache) error {
	return nil
}
func (r *activityUserCacheRepo) Delete(_ context.Context, _ string) error   { return nil }
func (r *activityUserCacheRepo) DeleteAll(_ context.Context) (int64, error) { return 0, nil }
func (r *activityUserCacheRepo) GetSelfBySessionFingerprint(_ context.Context, _ string) (*identity.UserCache, error) {
	return nil, nil
}
func (r *activityUserCacheRepo) UpsertSelf(_ context.Context, _ *identity.UserCache) error {
	return nil
}
func (r *activityUserCacheRepo) DeleteSelfRows(_ context.Context) error { return nil }

func (r *activityUserCacheRepo) ListContactsNeedingProfileResolution(_ context.Context) ([]*identity.UserCache, error) {
	return nil, nil
}

type worldInfoLookupRepo struct {
	byWorld map[string]*activity.WorldInfo
	calls   []string
}

func (w *worldInfoLookupRepo) UpsertVisit(_ context.Context, worldID string, _ time.Time) error {
	w.calls = append(w.calls, "visit:"+worldID)
	return nil
}

func (w *worldInfoLookupRepo) UpsertDisplayName(_ context.Context, worldID, _ string, _ time.Time) error {
	w.calls = append(w.calls, "name:"+worldID)
	return nil
}

func (w *worldInfoLookupRepo) GetByWorldID(_ context.Context, worldID string) (*activity.WorldInfo, error) {
	if w.byWorld == nil {
		return nil, nil
	}
	return w.byWorld[worldID], nil
}

func newActivityUC(t *testing.T) (*ActivityUseCase, *fakePlaySessionRepo, *memEncounterRepo, *fakeAppSettingsRepo) {
	t.Helper()
	play := &fakePlaySessionRepo{}
	enc := &memEncounterRepo{}
	settings := &fakeAppSettingsRepo{m: make(map[string]string)}
	return NewActivityUseCase(play, enc, settings, nil, nil), play, enc, settings
}

func TestActivityUseCase_RecordEncounterAt_joinLeaveAndUserCache(t *testing.T) {
	ctx := context.Background()
	enc := &memEncounterRepo{}
	users := &activityUserCacheRepo{}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, enc, &fakeAppSettingsRepo{m: make(map[string]string)}, users, nil)

	at := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	if err := uc.RecordEncounterAt(ctx, "", "usr_a", "Alice", activity.EncounterActionJoin, "wrld_x:1~abc", "", at); err != nil {
		t.Fatal(err)
	}
	if len(enc.encounters) != 1 || enc.encounters[0].WorldID == "" {
		t.Fatalf("join encounter = %+v", enc.encounters)
	}
	if users.byID["usr_a"] == nil || users.byID["usr_a"].DisplayName != "Alice" {
		t.Fatalf("user cache = %+v", users.byID["usr_a"])
	}

	leaveAt := at.Add(time.Hour)
	if err := uc.RecordEncounterAt(ctx, "", "usr_a", "Alice", activity.EncounterActionLeave, "", "", leaveAt); err != nil {
		t.Fatal(err)
	}
	if enc.encounters[0].LeftAt == nil || !enc.encounters[0].LeftAt.Equal(leaveAt) {
		t.Fatalf("LeftAt = %v, want %v", enc.encounters[0].LeftAt, leaveAt)
	}
}

type errActivityUserCacheRepo struct {
	activityUserCacheRepo
	saveErr error
	getErr  error
}

func (e *errActivityUserCacheRepo) Save(_ context.Context, u *identity.UserCache) error {
	if e.saveErr != nil {
		return e.saveErr
	}
	return e.activityUserCacheRepo.Save(context.Background(), u)
}

func (e *errActivityUserCacheRepo) GetByVRCUserID(_ context.Context, id string) (*identity.UserCache, error) {
	if e.getErr != nil {
		return nil, e.getErr
	}
	return e.activityUserCacheRepo.GetByVRCUserID(context.Background(), id)
}

func TestActivityUseCase_RecordEncounterAt_userCacheGetError(t *testing.T) {
	users := &errActivityUserCacheRepo{getErr: errors.New("read failed")}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, &memEncounterRepo{}, &fakeAppSettingsRepo{m: make(map[string]string)}, users, nil)
	err := uc.RecordEncounterAt(context.Background(), "", "usr_x", "X", activity.EncounterActionJoin, "inst", "", time.Now())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestActivityUseCase_RecordEncounterAt_userCacheSaveError(t *testing.T) {
	users := &errActivityUserCacheRepo{saveErr: errors.New("cache write failed")}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, &memEncounterRepo{}, &fakeAppSettingsRepo{m: make(map[string]string)}, users, nil)
	err := uc.RecordEncounterAt(context.Background(), "", "usr_x", "X", activity.EncounterActionJoin, "inst", "", time.Now())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestActivityUseCase_RecordEncounterAt_unknownActionNoOp(t *testing.T) {
	enc := &memEncounterRepo{}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, enc, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)
	if err := uc.RecordEncounterAt(context.Background(), "", "u", "U", "wave", "inst", "", time.Now()); err != nil {
		t.Fatal(err)
	}
	if len(enc.encounters) != 0 {
		t.Fatal("unknown action should not save")
	}
}

func TestActivityUseCase_RecordEncounter_delegatesToAt(t *testing.T) {
	enc := &memEncounterRepo{}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, enc, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)
	if err := uc.RecordEncounter(context.Background(), "usr_b", "Bob", activity.EncounterActionJoin, "wrld:1"); err != nil {
		t.Fatal(err)
	}
	if len(enc.encounters) != 1 {
		t.Fatalf("encounters = %d", len(enc.encounters))
	}
}

func TestActivityUseCase_ApplyCommand_recordEncounterJoin(t *testing.T) {
	enc := &memEncounterRepo{}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, enc, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)
	at := time.Date(2026, 3, 22, 12, 0, 0, 0, time.UTC)
	cmd := activity.RecordEncounterJoinCmd{
		VRCUserID: "usr_x", DisplayName: "X", InstanceID: "wrld_a:1", WorldID: "wrld_a", At: at,
	}
	if err := uc.ApplyCommand(context.Background(), "", cmd); err != nil {
		t.Fatal(err)
	}
	if len(enc.encounters) != 1 || enc.encounters[0].VRCUserID != "usr_x" {
		t.Fatalf("encounters = %+v", enc.encounters)
	}
}

func TestActivityUseCase_Checkpoint_roundtripAndClear(t *testing.T) {
	ctx := context.Background()
	uc, _, _, settings := newActivityUC(t)

	cp := &ActivityLogCheckpoint{WatchPath: "/logs", File: "output_log.txt", ByteOffset: 42}
	if err := uc.SetActivityLogCheckpoint(ctx, cp); err != nil {
		t.Fatal(err)
	}
	got, err := uc.GetActivityLogCheckpoint(ctx)
	if err != nil || got == nil {
		t.Fatalf("GetActivityLogCheckpoint = %+v err=%v", got, err)
	}
	fc, ok := got.FileCheckpoint("output_log.txt")
	if !ok || fc.ByteOffset != 42 {
		t.Fatalf("FileCheckpoint = %+v ok=%v, want ByteOffset 42", fc, ok)
	}
	if err := uc.SetActivityLogCheckpoint(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if settings.m["activity_log_checkpoint"] != "" {
		t.Fatalf("clear checkpoint: stored %q", settings.m["activity_log_checkpoint"])
	}
}

func TestActivityUseCase_GetActivityLogCheckpoint_invalidJSON(t *testing.T) {
	ctx := context.Background()
	settings := &fakeAppSettingsRepo{m: map[string]string{"activity_log_checkpoint": "{"}}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, &memEncounterRepo{}, settings, nil, nil)
	if _, err := uc.GetActivityLogCheckpoint(ctx); err == nil {
		t.Fatal("expected json error")
	}
}

func TestActivityUseCase_PlaySessionLifecycle(t *testing.T) {
	ctx := context.Background()
	const logSource = "/logs/output_log.txt"
	uc, play, _, _ := newActivityUC(t)

	start := time.Date(2025, 1, 10, 20, 0, 0, 0, time.UTC)
	if err := uc.StartPlaySession(ctx, logSource, "wrld:1~x", start); err != nil {
		t.Fatal(err)
	}
	open, _ := play.FindOpenForLogSource(ctx, logSource)
	if open == nil {
		t.Fatal("expected open session")
	}

	end := start.Add(2 * time.Hour)
	if err := uc.EndPlaySession(ctx, logSource, end); err != nil {
		t.Fatal(err)
	}
	closed, _ := play.FindOpenForLogSource(ctx, logSource)
	if closed != nil {
		t.Fatal("session should be closed")
	}
	if len(play.sessions) != 1 || play.sessions[0].EndTime == nil || *play.sessions[0].DurationSec != 7200 {
		t.Fatalf("closed session = %+v", play.sessions[0])
	}

	saved := &activity.PlaySession{StartTime: start, EndTime: &end}
	if err := uc.SavePlaySession(ctx, saved); err != nil {
		t.Fatal(err)
	}
	if saved.ID == "" {
		t.Fatal("SavePlaySession should assign ID when empty")
	}

	from := start.Add(-time.Hour)
	to := end.Add(time.Hour)
	list, err := uc.ListPlaySessions(ctx, from, to)
	if err != nil || len(list) < 2 {
		t.Fatalf("ListPlaySessions = %d err=%v", len(list), err)
	}
}

func TestActivityUseCase_GetActivityStats_and_parseDateRangeError(t *testing.T) {
	ctx := context.Background()
	play := &fakePlaySessionRepo{}
	start := time.Date(2025, 3, 1, 22, 0, 0, 0, time.Local)
	end := start.Add(3 * time.Hour)
	_ = play.Save(ctx, &activity.PlaySession{ID: "s", StartTime: start, EndTime: &end, DurationSec: intPtr(10800)})
	uc := NewActivityUseCase(play, &memEncounterRepo{}, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)

	stats, err := uc.GetActivityStats(ctx, "2025-03-01", "2025-03-02")
	if err != nil || stats == nil {
		t.Fatalf("GetActivityStats: %+v err=%v", stats, err)
	}
	if len(stats.DailyPlaySeconds) == 0 {
		t.Fatal("expected daily stats")
	}

	if _, err := uc.GetActivityStats(ctx, "bad", "2025-03-01"); err == nil {
		t.Fatal("expected parse error")
	}
	if _, err := uc.GetActivityStats(ctx, "2025-03-01", "bad"); err == nil {
		t.Fatal("expected parse error for toISO")
	}
}

func TestActivityUseCase_GetActivityStats_openSessionUsesCheckpoint(t *testing.T) {
	ctx := context.Background()
	play := &fakePlaySessionRepo{}
	start := time.Date(2025, 3, 2, 10, 0, 0, 0, time.Local)
	observed := time.Date(2025, 3, 2, 11, 0, 0, 0, time.Local)
	_ = play.Save(ctx, &activity.PlaySession{ID: "open", StartTime: start})
	settings := &fakeAppSettingsRepo{m: map[string]string{
		"activity_log_checkpoint": `{"watchPath":"/l","file":"output_log.txt","byteOffset":1,"vrChatLineTime":"` + observed.Format(time.RFC3339) + `"}`,
	}}
	uc := NewActivityUseCase(play, &memEncounterRepo{}, settings, nil, nil)

	stats, err := uc.GetActivityStats(ctx, "2025-03-02", "2025-03-02")
	if err != nil {
		t.Fatalf("GetActivityStats: %v", err)
	}
	if len(stats.DailyPlaySeconds) != 1 || stats.DailyPlaySeconds[0].Seconds != 3600 {
		t.Fatalf("daily = %+v", stats.DailyPlaySeconds)
	}
}

func TestActivityUseCase_GetActivityStats_excludesStaleOpenSessions(t *testing.T) {
	ctx := context.Background()
	play := &fakePlaySessionRepo{}
	staleStart := time.Date(2024, 1, 1, 10, 0, 0, 0, time.Local)
	recentStart := time.Date(2025, 3, 1, 22, 0, 0, 0, time.Local)
	recentEnd := recentStart.Add(2 * time.Hour)
	_ = play.Save(ctx, &activity.PlaySession{ID: "stale-open", StartTime: staleStart})
	_ = play.Save(ctx, &activity.PlaySession{
		ID: "recent-closed", StartTime: recentStart, EndTime: &recentEnd, DurationSec: intPtr(7200),
	})
	uc := NewActivityUseCase(play, &memEncounterRepo{}, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)

	stats, err := uc.GetActivityStats(ctx, "2025-03-01", "2025-03-02")
	if err != nil {
		t.Fatalf("GetActivityStats: %v", err)
	}
	var total int
	for _, d := range stats.DailyPlaySeconds {
		total += d.Seconds
	}
	if total != 7200 {
		t.Fatalf("total seconds = %d, want 7200 without stale open session bleed", total)
	}
}

func TestActivityUseCase_SetActivityLogFileCheckpoint_concurrent(t *testing.T) {
	ctx := context.Background()
	settings := &fakeAppSettingsRepo{m: make(map[string]string)}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, &memEncounterRepo{}, settings, nil, nil)

	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			path := "/logs/output_log_" + strconv.Itoa(n%4) + ".txt"
			_ = uc.SetActivityLogFileCheckpoint(ctx, "/watch", path, int64(n), "")
		}(i)
	}
	wg.Wait()

	cp, err := uc.GetActivityLogCheckpoint(ctx)
	if err != nil || cp == nil {
		t.Fatalf("checkpoint = %+v err=%v", cp, err)
	}
	if len(cp.Files) != 4 {
		t.Fatalf("files = %d, want 4 distinct log sources", len(cp.Files))
	}
}

func TestActivityUseCase_CloseOpenPlaySessionAtLastLogLine_lastLineBeforeStart(t *testing.T) {
	ctx := context.Background()
	const logSource = "/logs/output_log.txt"
	repo := &fakePlaySessionRepo{}
	start := time.Date(2024, 3, 18, 20, 0, 0, 0, time.Local)
	_ = repo.Save(ctx, &activity.PlaySession{ID: "s1", StartTime: start, LogSourcePath: logSource})
	uc := NewActivityUseCase(repo, stubEncounterRepo{}, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)
	last := start.Add(-time.Hour)
	if err := uc.CloseOpenPlaySessionAtLastLogLine(ctx, logSource, last); err != nil {
		t.Fatal(err)
	}
	open, _ := repo.FindOpenForLogSource(ctx, logSource)
	if open == nil {
		t.Fatal("session should remain open when lastLine before start")
	}
}

func TestActivityUseCase_GetActivityLogCheckpoint_missing(t *testing.T) {
	uc, _, _, _ := newActivityUC(t)
	got, err := uc.GetActivityLogCheckpoint(context.Background())
	if err != nil || got != nil {
		t.Fatalf("got = %+v err=%v", got, err)
	}
}

func TestActivityUseCase_CloseOpenPlaySessionAtLastLogLine_zeroTime(t *testing.T) {
	uc, _, _, _ := newActivityUC(t)
	if err := uc.CloseOpenPlaySessionAtLastLogLine(context.Background(), "/logs/output_log.txt", time.Time{}); err != nil {
		t.Fatal(err)
	}
}

func TestActivityUseCase_EndPlaySession_noOpenSession(t *testing.T) {
	uc, _, _, _ := newActivityUC(t)
	if err := uc.EndPlaySession(context.Background(), "/logs/output_log.txt", time.Now()); err != nil {
		t.Fatalf("EndPlaySession with no open session: %v", err)
	}
}

func intPtr(n int) *int { return &n }

func TestActivityUseCase_IsActivityDatastoreEmpty(t *testing.T) {
	ctx := context.Background()
	play := &fakePlaySessionRepo{}
	enc := &memEncounterRepo{}
	uc := NewActivityUseCase(play, enc, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)

	empty, err := uc.IsActivityDatastoreEmpty(ctx)
	if err != nil || !empty {
		t.Fatalf("empty = %v err=%v", empty, err)
	}
	_ = play.Save(ctx, &activity.PlaySession{ID: "x", StartTime: time.Now()})
	empty, err = uc.IsActivityDatastoreEmpty(ctx)
	if err != nil || empty {
		t.Fatalf("after play session empty = %v err=%v", empty, err)
	}
}

func TestActivityUseCase_WorldUpsertsAndBackfill(t *testing.T) {
	ctx := context.Background()
	world := &worldInfoLookupRepo{byWorld: map[string]*activity.WorldInfo{}}
	enc := &memEncounterRepo{backfillN: 3}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, enc, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, world)
	at := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)

	if err := uc.UpsertWorldVisit(ctx, "wrld_v", at); err != nil {
		t.Fatal(err)
	}
	if err := uc.UpsertWorldRoomName(ctx, "wrld_v", "Room", at); err != nil {
		t.Fatal(err)
	}
	if err := uc.UpsertWorldVisit(ctx, "", at); err != nil {
		t.Fatal(err)
	}

	n, err := uc.BackfillEncounterWorldContext(ctx)
	if err != nil || n != 3 {
		t.Fatalf("BackfillEncounterWorldContext = %d err=%v", n, err)
	}
}

func TestActivityUseCase_RotateEncounters_respectsRetentionSetting(t *testing.T) {
	ctx := context.Background()
	settings := &fakeAppSettingsRepo{m: map[string]string{"log_retention_days": "7"}}
	enc := &memEncounterRepo{
		encounters: []*activity.UserEncounter{
			{VRCUserID: "old", JoinedAt: time.Now().UTC().AddDate(0, 0, -30)},
			{VRCUserID: "new", JoinedAt: time.Now().UTC()},
		},
	}
	play := &fakePlaySessionRepo{
		sessions: []*activity.PlaySession{
			{ID: "old", StartTime: time.Now().UTC().AddDate(0, 0, -30)},
			{ID: "new", StartTime: time.Now().UTC()},
		},
	}
	uc := NewActivityUseCase(play, enc, settings, nil, nil)
	n, err := uc.RotateEncounters(ctx)
	if err != nil || n != 2 || len(enc.encounters) != 1 || len(play.sessions) != 1 {
		t.Fatalf("RotateEncounters n=%d enc=%d play=%d err=%v", n, len(enc.encounters), len(play.sessions), err)
	}
}

func TestActivityUseCase_RotateEncounters_settingsError(t *testing.T) {
	settings := &errOnGetSettingsRepo{fakeAppSettingsRepo: fakeAppSettingsRepo{m: make(map[string]string)}, getErr: errors.New("settings fail")}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, &memEncounterRepo{}, settings, nil, nil)
	_, err := uc.RotateEncounters(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestActivityUseCase_ListEncountersWithContext_enrichesWorldDisplayName(t *testing.T) {
	ctx := context.Background()
	world := &worldInfoLookupRepo{
		byWorld: map[string]*activity.WorldInfo{
			"wrld_cached": {WorldID: "wrld_cached", DisplayName: "Cached World"},
		},
	}
	enc := &memEncounterRepo{
		contexts: []*activity.EncounterWithContext{
			{
				Encounter: &activity.UserEncounter{VRCUserID: "u1", WorldID: "wrld_cached"},
			},
		},
	}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, enc, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, world)
	rows, err := uc.ListEncountersWithContext(ctx, nil)
	if err != nil || len(rows) != 1 || rows[0].WorldDisplayName != "Cached World" {
		t.Fatalf("rows = %+v err=%v", rows, err)
	}
}

func TestActivityUseCase_ListEncounters_delegates(t *testing.T) {
	enc := &memEncounterRepo{encounters: []*activity.UserEncounter{{VRCUserID: "u"}}}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, enc, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)
	got, err := uc.ListEncounters(context.Background(), nil)
	if err != nil || len(got) != 1 {
		t.Fatalf("ListEncounters = %+v err=%v", got, err)
	}
}

func TestActivityUseCase_CloseOpenEncountersAt_zeroTimeNoOp(t *testing.T) {
	rec := &recordingEncounterRepo{}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, rec, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)
	const logSource = "/logs/output_log.txt"
	if err := uc.CloseOpenEncountersAt(context.Background(), logSource, time.Time{}); err != nil {
		t.Fatal(err)
	}
	if len(rec.closeForLogSource) != 0 {
		t.Fatal("zero time should not call repo")
	}
	at := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	if err := uc.CloseOpenEncountersAt(context.Background(), logSource, at); err != nil {
		t.Fatal(err)
	}
	if len(rec.closeForLogSource) != 1 || rec.closeForLogSource[0].logSource != logSource || !rec.closeForLogSource[0].at.Equal(at) {
		t.Fatalf("calls = %v", rec.closeForLogSource)
	}
}

func TestActivityUseCase_ListEncountersWithContext_skipsWhenDisplayNameSet(t *testing.T) {
	world := &worldInfoLookupRepo{byWorld: map[string]*activity.WorldInfo{"wrld_x": {DisplayName: "Unused"}}}
	enc := &memEncounterRepo{
		contexts: []*activity.EncounterWithContext{{
			Encounter:        &activity.UserEncounter{WorldID: "wrld_x"},
			WorldDisplayName: "Already Set",
		}},
	}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, enc, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, world)
	rows, err := uc.ListEncountersWithContext(context.Background(), nil)
	if err != nil || rows[0].WorldDisplayName != "Already Set" {
		t.Fatalf("rows = %+v err=%v", rows, err)
	}
}

func TestActivityUseCase_ListEncountersWithContext_worldFromInstanceKey(t *testing.T) {
	world := &worldInfoLookupRepo{
		byWorld: map[string]*activity.WorldInfo{
			"wrld_inst": {WorldID: "wrld_inst", DisplayName: "From Instance"},
		},
	}
	enc := &memEncounterRepo{
		contexts: []*activity.EncounterWithContext{{
			Encounter: &activity.UserEncounter{InstanceID: "wrld_inst:1~abc"},
		}},
	}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, enc, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, world)
	rows, err := uc.ListEncountersWithContext(context.Background(), nil)
	if err != nil || rows[0].WorldDisplayName != "From Instance" {
		t.Fatalf("rows = %+v err=%v", rows, err)
	}
}

type countingPlaySessionRepo struct {
	fakePlaySessionRepo
	countErr error
}

func (c *countingPlaySessionRepo) Count(_ context.Context) (int64, error) {
	if c.countErr != nil {
		return 0, c.countErr
	}
	return c.fakePlaySessionRepo.Count(context.Background())
}

func TestActivityUseCase_UpsertWorldRoomName_emptyNoOp(t *testing.T) {
	world := &worldInfoLookupRepo{}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, &memEncounterRepo{}, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, world)
	at := time.Now()
	if err := uc.UpsertWorldRoomName(context.Background(), "wrld_x", "", at); err != nil {
		t.Fatal(err)
	}
	if len(world.calls) != 0 {
		t.Fatalf("calls = %v", world.calls)
	}
}

func TestActivityUseCase_ListEncountersWithContext_nilWorldRepo(t *testing.T) {
	enc := &memEncounterRepo{
		contexts: []*activity.EncounterWithContext{{Encounter: &activity.UserEncounter{VRCUserID: "u"}}},
	}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, enc, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)
	rows, err := uc.ListEncountersWithContext(context.Background(), nil)
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows = %+v err=%v", rows, err)
	}
}

func TestActivityUseCase_IsActivityDatastoreEmpty_countError(t *testing.T) {
	play := &countingPlaySessionRepo{countErr: errors.New("db down")}
	uc := NewActivityUseCase(play, &memEncounterRepo{}, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)
	_, err := uc.IsActivityDatastoreEmpty(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestActivityUseCase_RecordEncounterAt_idempotentJoin(t *testing.T) {
	ctx := context.Background()
	enc := &memEncounterRepo{}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, enc, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)
	at := time.Date(2026, 6, 22, 22, 51, 17, 0, time.UTC)
	inst := "wrld_a:1~region(jp)"
	for i := 0; i < 2; i++ {
		if err := uc.RecordEncounterAt(ctx, "", "usr_a", "User A", activity.EncounterActionJoin, inst, "wrld_a", at); err != nil {
			t.Fatal(err)
		}
	}
	if len(enc.encounters) != 1 {
		t.Fatalf("encounters = %d, want 1", len(enc.encounters))
	}
	if enc.encounters[0].InstanceID != inst {
		t.Fatalf("instance_id = %q, want %q", enc.encounters[0].InstanceID, inst)
	}
}

func TestActivityUseCase_CloseOpenEncountersAt_skipsFutureJoins(t *testing.T) {
	ctx := context.Background()
	const logSource = "/logs/output_log.txt"
	enc := &memEncounterRepo{}
	uc := NewActivityUseCase(&fakePlaySessionRepo{}, enc, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)
	early := time.Date(2026, 6, 22, 18, 45, 54, 0, time.UTC)
	lateJoin := time.Date(2026, 6, 22, 22, 51, 17, 0, time.UTC)
	if err := uc.RecordEncounterAt(ctx, logSource, "usr_a", "User A", activity.EncounterActionJoin, "wrld_b:1", "wrld_b", lateJoin); err != nil {
		t.Fatal(err)
	}
	if err := uc.CloseOpenEncountersAt(ctx, logSource, early); err != nil {
		t.Fatal(err)
	}
	if enc.encounters[0].LeftAt != nil {
		t.Fatalf("future join should stay open, left_at=%v", enc.encounters[0].LeftAt)
	}
}

func TestActivityUseCase_logSourceScopedClose(t *testing.T) {
	ctx := context.Background()
	const logA = "/logs/output_log_a.txt"
	const logB = "/logs/output_log_b.txt"
	play := &fakePlaySessionRepo{}
	enc := &memEncounterRepo{}
	uc := NewActivityUseCase(play, enc, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)

	startA := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)
	startB := time.Date(2026, 3, 21, 11, 0, 0, 0, time.UTC)
	closeAt := time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC)

	if err := uc.StartPlaySession(ctx, logA, "wrld_a:1", startA); err != nil {
		t.Fatal(err)
	}
	if err := uc.StartPlaySession(ctx, logB, "wrld_b:1", startB); err != nil {
		t.Fatal(err)
	}
	if err := enc.Save(ctx, &activity.UserEncounter{
		VRCUserID: "usr_a", DisplayName: "A", JoinedAt: startA, LogSourcePath: logA,
	}); err != nil {
		t.Fatal(err)
	}
	if err := enc.Save(ctx, &activity.UserEncounter{
		VRCUserID: "usr_b", DisplayName: "B", JoinedAt: startB, LogSourcePath: logB,
	}); err != nil {
		t.Fatal(err)
	}

	if err := uc.FinalizeOpenActivityForLogSource(ctx, logA, closeAt); err != nil {
		t.Fatal(err)
	}

	openA, _ := play.FindOpenForLogSource(ctx, logA)
	if openA != nil {
		t.Fatal("log A session should be closed")
	}
	openB, _ := play.FindOpenForLogSource(ctx, logB)
	if openB == nil {
		t.Fatal("log B session should remain open")
	}
	for _, e := range enc.encounters {
		switch e.VRCUserID {
		case "usr_a":
			if e.LeftAt == nil || !e.LeftAt.Equal(closeAt) {
				t.Fatalf("usr_a left_at = %v, want %v", e.LeftAt, closeAt)
			}
		case "usr_b":
			if e.LeftAt != nil {
				t.Fatalf("usr_b should stay open, left_at = %v", e.LeftAt)
			}
		}
	}
}
