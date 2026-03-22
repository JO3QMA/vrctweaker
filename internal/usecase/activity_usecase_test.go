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

func (f *fakePlaySessionRepo) List(context.Context, time.Time, time.Time) ([]*activity.PlaySession, error) {
	return nil, nil
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

func TestActivityUseCase_CloseOpenPlaySessionAtLastLogLineIfSameLocalDay_closesWhenSameDay(t *testing.T) {
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
	if err := uc.CloseOpenPlaySessionAtLastLogLineIfSameLocalDay(ctx, lastLine); err != nil {
		t.Fatalf("CloseOpenPlaySessionAtLastLogLineIfSameLocalDay: %v", err)
	}
	open, err := repo.FindLatestWithoutEndTime(ctx)
	if err != nil {
		t.Fatalf("FindLatestWithoutEndTime: %v", err)
	}
	if open != nil {
		t.Fatalf("expected session closed, still open: %+v", open)
	}
}

func TestActivityUseCase_CloseOpenPlaySessionAtLastLogLineIfSameLocalDay_skipsWhenCrossesMidnight(t *testing.T) {
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
	if err := uc.CloseOpenPlaySessionAtLastLogLineIfSameLocalDay(ctx, lastLine); err != nil {
		t.Fatalf("CloseOpenPlaySessionAtLastLogLineIfSameLocalDay: %v", err)
	}
	open, err := repo.FindLatestWithoutEndTime(ctx)
	if err != nil {
		t.Fatalf("FindLatestWithoutEndTime: %v", err)
	}
	if open == nil {
		t.Fatal("expected session still open when last line is next calendar day")
	}
}
