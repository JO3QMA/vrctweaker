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

func (stubEncounterRepo) DeleteOlderThan(context.Context, time.Time) (int64, error) { return 0, nil }

func (stubEncounterRepo) DeleteAll(context.Context) (int64, error) { return 0, nil }

func (stubEncounterRepo) Count(context.Context) (int64, error) { return 0, nil }

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

func TestActivityEventHandler_RoomNameAfterOnLeftRoom_usesPendingDestinationWorld(t *testing.T) {
	ctx := context.Background()
	base := time.Date(2026, 3, 22, 11, 23, 51, 0, time.UTC)
	spy := &spyWorldInfoRepo{}
	uc := usecase.NewActivityUseCase(stubPlaySessionRepo{}, stubEncounterRepo{}, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, spy)
	h := NewActivityEventHandler(uc, ctx, nil, nil)

	h.Handle(&activity.DestinationSetEvent{
		WorldID:      testWorldID,
		FullInstance: testFullInstance,
		OccurredAt:   base,
	})
	h.Handle(&activity.SessionEvent{Type: activity.SessionEventEnd, OccurredAt: base})
	h.Handle(&activity.RoomNameEvent{RoomName: "SuRroom", OccurredAt: base})
	h.Handle(&activity.SessionEvent{Type: activity.SessionEventStart, InstanceID: testFullInstance, OccurredAt: base})

	if len(spy.displayNameCalls) != 1 {
		t.Fatalf("UpsertDisplayName calls = %d, want 1: %+v", len(spy.displayNameCalls), spy.displayNameCalls)
	}
	if spy.displayNameCalls[0].worldID != testWorldID || spy.displayNameCalls[0].displayName != "SuRroom" {
		t.Errorf("UpsertDisplayName = %+v, want world %q name %q", spy.displayNameCalls[0], testWorldID, "SuRroom")
	}
}

func TestActivityEventHandler_RoomNameWithoutOnLeftRoom_unchanged(t *testing.T) {
	ctx := context.Background()
	base := time.Date(2026, 3, 22, 11, 22, 51, 0, time.UTC)
	const homeWorld = "wrld_e055f1a3-6fcb-4d19-9945-f0a1c92cc19b"
	homeInst := homeWorld + ":04910~private(usr_dec48a78-894a-4ef3-8524-8cf546ad1b2e)~region(jp)"
	spy := &spyWorldInfoRepo{}
	uc := usecase.NewActivityUseCase(stubPlaySessionRepo{}, stubEncounterRepo{}, &fakeAppSettingsRepo{m: make(map[string]string)}, nil, spy)
	h := NewActivityEventHandler(uc, ctx, nil, nil)

	h.Handle(&activity.DestinationSetEvent{
		WorldID:      homeWorld,
		FullInstance: homeInst,
		OccurredAt:   base,
	})
	h.Handle(&activity.RoomNameEvent{RoomName: "ホームチェックv6․0", OccurredAt: base})
	h.Handle(&activity.SessionEvent{Type: activity.SessionEventStart, InstanceID: homeInst, OccurredAt: base})

	if len(spy.displayNameCalls) != 1 {
		t.Fatalf("UpsertDisplayName calls = %d, want 1: %+v", len(spy.displayNameCalls), spy.displayNameCalls)
	}
	if spy.displayNameCalls[0].worldID != homeWorld || spy.displayNameCalls[0].displayName != "ホームチェックv6․0" {
		t.Errorf("UpsertDisplayName = %+v", spy.displayNameCalls[0])
	}
}
