package logwatcher

import (
	"context"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/settings"
	"vrchat-tweaker/internal/usecase"
)

const testWorldID = "wrld_beddab1e-fee1-cafe-f00d-ca7c0dd1eca7"

var testFullInstance = testWorldID + ":41550~hidden(usr_aeab2f4d-40b4-4f73-acbd-608ac47b763e)~region(jp)"

type spyWorldInfoRepo struct {
	displayNameCalls []struct {
		worldID     string
		displayName string
	}
}

func (s *spyWorldInfoRepo) UpsertVisit(context.Context, string, time.Time) error { return nil }

func (s *spyWorldInfoRepo) UpsertDisplayName(_ context.Context, worldID, displayName string, _ time.Time) error {
	s.displayNameCalls = append(s.displayNameCalls, struct {
		worldID     string
		displayName string
	}{worldID, displayName})
	return nil
}

func (s *spyWorldInfoRepo) GetByWorldID(context.Context, string) (*activity.WorldInfo, error) {
	return nil, nil
}

type stubPlaySessionRepo struct{}

func (stubPlaySessionRepo) List(context.Context, time.Time, time.Time) ([]*activity.PlaySession, error) {
	return nil, nil
}

func (stubPlaySessionRepo) GetByID(context.Context, string) (*activity.PlaySession, error) {
	return nil, nil
}

func (stubPlaySessionRepo) Save(context.Context, *activity.PlaySession) error { return nil }

func (stubPlaySessionRepo) FindLatestWithoutEndTime(context.Context) (*activity.PlaySession, error) {
	return nil, nil
}

func (stubPlaySessionRepo) Count(context.Context) (int64, error) { return 0, nil }

type stubEncounterRepo struct{}

func (stubEncounterRepo) List(context.Context, *activity.EncounterFilter) ([]*activity.UserEncounter, error) {
	return nil, nil
}

func (stubEncounterRepo) ListWithContext(context.Context, *activity.EncounterFilter) ([]*activity.EncounterWithContext, error) {
	return nil, nil
}

func (stubEncounterRepo) Save(context.Context, *activity.UserEncounter) error { return nil }

func (stubEncounterRepo) CloseEncounterLeave(context.Context, string, time.Time) (int64, error) {
	return 0, nil
}

func (stubEncounterRepo) CloseOpenEncountersAt(context.Context, time.Time) (int64, error) {
	return 0, nil
}

func (stubEncounterRepo) DeleteOlderThan(context.Context, time.Time) (int64, error) { return 0, nil }

func (stubEncounterRepo) DeleteAll(context.Context) (int64, error) { return 0, nil }

func (stubEncounterRepo) Count(context.Context) (int64, error) { return 0, nil }

func (stubEncounterRepo) BackfillMissingWorldContext(context.Context) (int64, error) { return 0, nil }

type spyEncounterRepo struct {
	stubEncounterRepo
	saves       []*activity.UserEncounter
	closeLeaves []struct {
		VRCUserID string
		At        time.Time
	}
}

func (s *spyEncounterRepo) Save(_ context.Context, e *activity.UserEncounter) error {
	c := *e
	s.saves = append(s.saves, &c)
	return nil
}

func (s *spyEncounterRepo) CloseEncounterLeave(_ context.Context, vrcUserID string, leftAt time.Time) (int64, error) {
	s.closeLeaves = append(s.closeLeaves, struct {
		VRCUserID string
		At        time.Time
	}{vrcUserID, leftAt})
	return 1, nil
}

type fakeAppSettingsRepo struct {
	m map[string]string
}

func (f *fakeAppSettingsRepo) Get(_ context.Context, key string) (string, error) {
	return f.m[key], nil
}

func (f *fakeAppSettingsRepo) Set(_ context.Context, key, value string) error {
	if f.m == nil {
		f.m = make(map[string]string)
	}
	f.m[key] = value
	return nil
}

func (f *fakeAppSettingsRepo) GetAll(_ context.Context) (map[string]string, error) {
	out := make(map[string]string, len(f.m))
	for k, v := range f.m {
		out[k] = v
	}
	return out, nil
}

var _ settings.AppSettingsRepository = (*fakeAppSettingsRepo)(nil)

func TestActivityIngestAdapter_SuppressEncounterNotify_skipsOnAfterEncounter(t *testing.T) {
	ctx := context.Background()
	base := time.Date(2026, 3, 22, 12, 0, 0, 0, time.UTC)
	var calls int
	cb := func() { calls++ }
	uc := usecase.NewActivityUseCase(stubPlaySessionRepo{}, stubEncounterRepo{}, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)
	a := NewActivityIngestAdapter(uc, ctx, nil, cb)
	a.Handle(&activity.SessionEvent{Type: activity.SessionEventStart, InstanceID: testFullInstance, OccurredAt: base})

	a.SetSuppressEncounterNotify(true)
	a.Handle(&activity.EncounterEvent{
		VRCUserID:     "usr_suppress_test",
		DisplayName:   "A",
		Action:        activity.EncounterActionJoin,
		EncounteredAt: base.Add(time.Second),
	})
	if calls != 0 {
		t.Errorf("onAfterEncounter calls = %d, want 0 while suppressed", calls)
	}

	a.SetSuppressEncounterNotify(false)
	a.Handle(&activity.EncounterEvent{
		VRCUserID:     "usr_suppress_test2",
		DisplayName:   "B",
		Action:        activity.EncounterActionJoin,
		EncounteredAt: base.Add(2 * time.Second),
	})
	if calls != 1 {
		t.Errorf("onAfterEncounter calls = %d, want 1 after unsuppress", calls)
	}
}

func TestActivityIngestAdapter_EndToEndEncounterPersistence(t *testing.T) {
	ctx := context.Background()
	base := time.Date(2026, 3, 18, 0, 1, 0, 0, time.UTC)
	const minasocoWorld = "wrld_c03f8195-3c64-46d8-b5ae-242f214c9404"
	minasocoInst := minasocoWorld + ":98225~hidden(usr_83ba5dc2-2912-4a21-a514-8b954e60a79b)~region(jp)"
	const otherUser = "usr_1564b5c1-888a-4d08-b7f4-dcedcf702a90"

	spy := &spyEncounterRepo{}
	uc := usecase.NewActivityUseCase(stubPlaySessionRepo{}, spy, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, nil)
	a := NewActivityIngestAdapter(uc, ctx, nil, nil)

	a.Handle(&activity.DestinationSetEvent{
		WorldID:      minasocoWorld,
		FullInstance: minasocoInst,
		OccurredAt:   base,
	})
	a.Handle(&activity.SessionEvent{Type: activity.SessionEventStart, InstanceID: minasocoInst, OccurredAt: base})
	a.Handle(&activity.EncounterEvent{
		VRCUserID:     otherUser,
		DisplayName:   "Nau_UoxoU",
		Action:        activity.EncounterActionJoin,
		EncounteredAt: base,
	})
	a.Handle(&activity.EncounterEvent{
		VRCUserID:     otherUser,
		DisplayName:   "Nau_UoxoU",
		Action:        activity.EncounterActionLeave,
		EncounteredAt: base.Add(time.Second),
	})

	if len(spy.saves) != 1 {
		t.Fatalf("Save calls = %d, want 1", len(spy.saves))
	}
	e := spy.saves[0]
	if e.WorldID != minasocoWorld || e.InstanceID != minasocoInst {
		t.Errorf("Save world_id=%q instance_id=%q, want world %q instance %q",
			e.WorldID, e.InstanceID, minasocoWorld, minasocoInst)
	}
	if len(spy.closeLeaves) != 1 || spy.closeLeaves[0].VRCUserID != otherUser {
		t.Fatalf("closeLeaves = %+v, want one leave for %s", spy.closeLeaves, otherUser)
	}
}
