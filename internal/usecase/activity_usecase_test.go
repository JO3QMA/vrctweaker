package usecase

import (
	"context"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
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

func (f *fakePlaySessionRepo) Count(context.Context) (int64, error) {
	return int64(len(f.sessions)), nil
}

type stubEncounterRepo struct{}

func (stubEncounterRepo) List(context.Context, *activity.EncounterFilter) ([]*activity.UserEncounter, error) {
	return nil, nil
}

func (stubEncounterRepo) ListWithContext(context.Context, *activity.EncounterFilter) ([]*activity.EncounterWithContext, error) {
	return nil, nil
}

func (stubEncounterRepo) Save(context.Context, *activity.UserEncounter) error { return nil }

func (stubEncounterRepo) DeleteOlderThan(context.Context, time.Time) (int64, error) { return 0, nil }

func (stubEncounterRepo) DeleteAll(context.Context) (int64, error) { return 0, nil }

func (stubEncounterRepo) Count(context.Context) (int64, error) { return 0, nil }

func TestActivityUseCase_CloseOpenPlaySessionAtLastLogLine_sameDay(t *testing.T) {
	ctx := context.Background()
	repo := &fakePlaySessionRepo{}
	uc := NewActivityUseCase(repo, stubEncounterRepo{}, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)

	start := time.Date(2024, 3, 18, 10, 0, 0, 0, time.Local)
	_ = repo.Save(ctx, &activity.PlaySession{
		ID:        "s1",
		StartTime: start,
		EndTime:   nil,
	})

	lastLine := time.Date(2024, 3, 18, 20, 0, 0, 0, time.Local)
	if err := uc.CloseOpenPlaySessionAtLastLogLine(ctx, lastLine); err != nil {
		t.Fatalf("CloseOpenPlaySessionAtLastLogLine: %v", err)
	}
	open, err := repo.FindLatestWithoutEndTime(ctx)
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
	repo := &fakePlaySessionRepo{}
	uc := NewActivityUseCase(repo, stubEncounterRepo{}, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)

	start := time.Date(2024, 3, 18, 23, 0, 0, 0, time.Local)
	_ = repo.Save(ctx, &activity.PlaySession{
		ID:        "s1",
		StartTime: start,
		EndTime:   nil,
	})

	lastLine := time.Date(2024, 3, 19, 1, 0, 0, 0, time.Local)
	if err := uc.CloseOpenPlaySessionAtLastLogLine(ctx, lastLine); err != nil {
		t.Fatalf("CloseOpenPlaySessionAtLastLogLine: %v", err)
	}
	open, err := repo.FindLatestWithoutEndTime(ctx)
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
	repo := &fakePlaySessionRepo{}
	uc := NewActivityUseCase(repo, stubEncounterRepo{}, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)

	start := time.Date(2024, 3, 18, 20, 0, 0, 0, time.Local)
	_ = repo.Save(ctx, &activity.PlaySession{
		ID:        "s1",
		StartTime: start,
		EndTime:   nil,
	})

	lastLine := time.Date(2024, 3, 20, 3, 0, 0, 0, time.Local)
	if err := uc.CloseOpenPlaySessionAtLastLogLine(ctx, lastLine); err != nil {
		t.Fatalf("CloseOpenPlaySessionAtLastLogLine: %v", err)
	}
	open3, err3 := repo.FindLatestWithoutEndTime(ctx)
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
