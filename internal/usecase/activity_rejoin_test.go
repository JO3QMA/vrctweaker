package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
)

const testRejoinInst = "wrld_test1111-1111-4111-8111-111111111101:42~public"

type fakeRejoinPlayRepo struct {
	sessions []*activity.PlaySession
	err      error
}

func (f *fakeRejoinPlayRepo) List(context.Context, time.Time, time.Time) ([]*activity.PlaySession, error) {
	return f.sessions, f.err
}
func (f *fakeRejoinPlayRepo) GetByID(context.Context, string) (*activity.PlaySession, error) {
	return nil, nil
}
func (f *fakeRejoinPlayRepo) Save(context.Context, *activity.PlaySession) error { return nil }
func (f *fakeRejoinPlayRepo) FindLatestWithoutEndTime(context.Context) (*activity.PlaySession, error) {
	return nil, nil
}
func (f *fakeRejoinPlayRepo) FindLatestWithInstanceID(context.Context) (*activity.PlaySession, error) {
	if f.err != nil {
		return nil, f.err
	}
	var best *activity.PlaySession
	for _, s := range f.sessions {
		if s == nil || s.InstanceID == "" {
			continue
		}
		if best == nil || s.StartTime.After(best.StartTime) {
			best = s
		}
	}
	return best, nil
}
func (f *fakeRejoinPlayRepo) FindOpenForLogSource(context.Context, string) (*activity.PlaySession, error) {
	return nil, nil
}
func (f *fakeRejoinPlayRepo) FindAllWithoutEndTime(context.Context) ([]*activity.PlaySession, error) {
	return nil, nil
}
func (f *fakeRejoinPlayRepo) Count(context.Context) (int64, error) { return 0, nil }
func (f *fakeRejoinPlayRepo) DeleteOlderThan(context.Context, time.Time) (int64, error) {
	return 0, nil
}

type fakeRejoinWorldRepo struct {
	names map[string]string
}

func (f *fakeRejoinWorldRepo) UpsertVisit(context.Context, string, time.Time) error { return nil }
func (f *fakeRejoinWorldRepo) UpsertDisplayName(context.Context, string, string, time.Time) error {
	return nil
}
func (f *fakeRejoinWorldRepo) GetByWorldID(_ context.Context, worldID string) (*activity.WorldInfo, error) {
	if f.names == nil {
		return nil, nil
	}
	name, ok := f.names[worldID]
	if !ok || name == "" {
		return nil, nil
	}
	return &activity.WorldInfo{WorldID: worldID, DisplayName: name}, nil
}

func TestGetRejoinTarget_noSessions(t *testing.T) {
	uc := NewActivityUseCase(&fakeRejoinPlayRepo{}, &memEncounterRepo{}, &fakeAppSettingsRepo{m: map[string]string{}}, nil, nil)
	got, err := uc.GetRejoinTarget(context.Background())
	if err != nil || got != nil {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestGetRejoinTarget_skipsEmptyInstanceID(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
	repo := &fakeRejoinPlayRepo{sessions: []*activity.PlaySession{
		{ID: "empty", StartTime: t1, InstanceID: ""},
		{ID: "good", StartTime: t0, InstanceID: testRejoinInst},
	}}
	uc := NewActivityUseCase(repo, &memEncounterRepo{}, &fakeAppSettingsRepo{m: map[string]string{}}, nil, nil)
	got, err := uc.GetRejoinTarget(context.Background())
	if err != nil || got == nil || got.InstanceID != testRejoinInst {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestGetRejoinTarget_picksLatestStartTime(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	repo := &fakeRejoinPlayRepo{sessions: []*activity.PlaySession{
		{ID: "old", StartTime: t0, InstanceID: "wrld_old:1~public"},
		{ID: "new", StartTime: t1, InstanceID: testRejoinInst},
	}}
	uc := NewActivityUseCase(repo, &memEncounterRepo{}, &fakeAppSettingsRepo{m: map[string]string{}}, nil, nil)
	got, err := uc.GetRejoinTarget(context.Background())
	if err != nil || got == nil || got.PlaySessionID != "new" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestGetRejoinTarget_withWorldDisplayName(t *testing.T) {
	wid := activity.WorldIDFromInstanceKey(testRejoinInst)
	repo := &fakeRejoinPlayRepo{sessions: []*activity.PlaySession{
		{ID: "s1", StartTime: time.Now().UTC(), InstanceID: testRejoinInst},
	}}
	world := &fakeRejoinWorldRepo{names: map[string]string{wid: "Test World"}}
	uc := NewActivityUseCase(repo, &memEncounterRepo{}, &fakeAppSettingsRepo{m: map[string]string{}}, nil, world)
	got, err := uc.GetRejoinTarget(context.Background())
	if err != nil || got == nil || got.WorldDisplayName != "Test World" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestGetRejoinTarget_dbError(t *testing.T) {
	repo := &fakeRejoinPlayRepo{err: errors.New("db down")}
	uc := NewActivityUseCase(repo, &memEncounterRepo{}, &fakeAppSettingsRepo{m: map[string]string{}}, nil, nil)
	got, err := uc.GetRejoinTarget(context.Background())
	if err == nil || got != nil {
		t.Fatalf("got=%v err=%v", got, err)
	}
}
