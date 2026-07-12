package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/launcher"
	"vrchat-tweaker/internal/usecase"
)

type errRejoinPlayRepo struct{}

func (errRejoinPlayRepo) List(context.Context, time.Time, time.Time) ([]*activity.PlaySession, error) {
	return nil, errors.New("db down")
}
func (errRejoinPlayRepo) GetByID(context.Context, string) (*activity.PlaySession, error) {
	return nil, errors.New("db down")
}
func (errRejoinPlayRepo) Save(context.Context, *activity.PlaySession) error { return errors.New("db down") }
func (errRejoinPlayRepo) FindLatestWithoutEndTime(context.Context) (*activity.PlaySession, error) {
	return nil, errors.New("db down")
}
func (errRejoinPlayRepo) FindLatestWithInstanceID(context.Context) (*activity.PlaySession, error) {
	return nil, errors.New("db down")
}
func (errRejoinPlayRepo) FindOpenForLogSource(context.Context, string) (*activity.PlaySession, error) {
	return nil, errors.New("db down")
}
func (errRejoinPlayRepo) FindAllWithoutEndTime(context.Context) ([]*activity.PlaySession, error) {
	return nil, errors.New("db down")
}
func (errRejoinPlayRepo) Count(context.Context) (int64, error) { return 0, errors.New("db down") }
func (errRejoinPlayRepo) DeleteOlderThan(context.Context, time.Time) (int64, error) {
	return 0, errors.New("db down")
}

func TestGetInstanceRejoinSection_degradesOnError(t *testing.T) {
	a := &App{ctx: context.Background()}
	a.activity = usecase.NewActivityUseCase(errRejoinPlayRepo{}, nil, &memSettingsRepo{m: map[string]string{}}, nil, nil)
	a.launcher = usecase.NewLauncherUseCase(&memLaunchRepo{profiles: []*launcher.LaunchProfile{{ID: "p1", Name: "P"}}})
	a.settings = usecase.NewSettingsUseCase(&memSettingsRepo{m: map[string]string{}})
	got, err := a.GetInstanceRejoinSection()
	if err != nil || got != nil {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestInstanceRejoin_emptyProfileID(t *testing.T) {
	a := &App{ctx: context.Background()}
	err := a.InstanceRejoin("")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetInstanceRejoinSection_noProfiles(t *testing.T) {
	a := &App{ctx: context.Background()}
	a.activity = usecase.NewActivityUseCase(&memRejoinPlayRepo{session: &activity.PlaySession{
		ID: "s1", StartTime: time.Now().UTC(), InstanceID: "wrld_x:1~public",
	}}, nil, &memSettingsRepo{m: map[string]string{}}, nil, nil)
	a.launcher = usecase.NewLauncherUseCase(&memLaunchRepo{})
	a.settings = usecase.NewSettingsUseCase(&memSettingsRepo{m: map[string]string{}})
	got, err := a.GetInstanceRejoinSection()
	if err != nil || got != nil {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestGetInstanceRejoinSection_withWorldDisplayName(t *testing.T) {
	inst := "wrld_test1111-1111-4111-8111-111111111101:42~public"
	wid := activity.WorldIDFromInstanceKey(inst)
	a := &App{ctx: context.Background()}
	a.activity = usecase.NewActivityUseCase(&memRejoinPlayRepo{session: &activity.PlaySession{
		ID: "s1", StartTime: time.Now().UTC(), InstanceID: inst,
	}}, nil, &memSettingsRepo{m: map[string]string{}}, nil, &memRejoinWorldRepo{names: map[string]string{wid: "Test World"}})
	a.launcher = usecase.NewLauncherUseCase(&memLaunchRepo{profiles: []*launcher.LaunchProfile{{ID: "p1", Name: "P"}}})
	a.settings = usecase.NewSettingsUseCase(&memSettingsRepo{m: map[string]string{}})
	got, err := a.GetInstanceRejoinSection()
	if err != nil || got == nil || got.WorldDisplayName != "Test World" || got.SelectedProfileID != "p1" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestInstanceRejoin_profileNotFound(t *testing.T) {
	a := &App{ctx: context.Background()}
	a.activity = usecase.NewActivityUseCase(&memRejoinPlayRepo{session: &activity.PlaySession{
		ID: "s1", StartTime: time.Now().UTC(), InstanceID: "wrld_x:1~public",
	}}, nil, &memSettingsRepo{m: map[string]string{}}, nil, nil)
	a.launcher = usecase.NewLauncherUseCase(&memLaunchRepo{profiles: []*launcher.LaunchProfile{{ID: "p1", Name: "P"}}})
	a.settings = usecase.NewSettingsUseCase(&memSettingsRepo{m: map[string]string{}})
	err := a.InstanceRejoin("missing")
	if err == nil {
		t.Fatal("expected error")
	}
}

type memRejoinPlayRepo struct {
	session *activity.PlaySession
}

func (r *memRejoinPlayRepo) List(context.Context, time.Time, time.Time) ([]*activity.PlaySession, error) {
	return nil, nil
}
func (r *memRejoinPlayRepo) GetByID(context.Context, string) (*activity.PlaySession, error) {
	return nil, nil
}
func (r *memRejoinPlayRepo) Save(context.Context, *activity.PlaySession) error { return nil }
func (r *memRejoinPlayRepo) FindLatestWithoutEndTime(context.Context) (*activity.PlaySession, error) {
	return nil, nil
}
func (r *memRejoinPlayRepo) FindLatestWithInstanceID(context.Context) (*activity.PlaySession, error) {
	return r.session, nil
}
func (r *memRejoinPlayRepo) FindOpenForLogSource(context.Context, string) (*activity.PlaySession, error) {
	return nil, nil
}
func (r *memRejoinPlayRepo) FindAllWithoutEndTime(context.Context) ([]*activity.PlaySession, error) {
	return nil, nil
}
func (r *memRejoinPlayRepo) Count(context.Context) (int64, error) { return 0, nil }
func (r *memRejoinPlayRepo) DeleteOlderThan(context.Context, time.Time) (int64, error) {
	return 0, nil
}

type memRejoinWorldRepo struct {
	names map[string]string
}

func (memRejoinWorldRepo) UpsertVisit(context.Context, string, time.Time) error { return nil }
func (memRejoinWorldRepo) UpsertDisplayName(context.Context, string, string, time.Time) error {
	return nil
}
func (r *memRejoinWorldRepo) GetByWorldID(_ context.Context, worldID string) (*activity.WorldInfo, error) {
	name, ok := r.names[worldID]
	if !ok || name == "" {
		return nil, nil
	}
	return &activity.WorldInfo{WorldID: worldID, DisplayName: name}, nil
}

type memSettingsRepo struct {
	m map[string]string
}

func (r *memSettingsRepo) Get(_ context.Context, key string) (string, error) {
	return r.m[key], nil
}
func (r *memSettingsRepo) Set(_ context.Context, key, value string) error {
	r.m[key] = value
	return nil
}
func (r *memSettingsRepo) GetAll(_ context.Context) (map[string]string, error) {
	out := make(map[string]string, len(r.m))
	for k, v := range r.m {
		out[k] = v
	}
	return out, nil
}

type memLaunchRepo struct {
	profiles []*launcher.LaunchProfile
}

func (r *memLaunchRepo) List(context.Context) ([]*launcher.LaunchProfile, error) { return r.profiles, nil }
func (r *memLaunchRepo) GetByID(_ context.Context, id string) (*launcher.LaunchProfile, error) {
	for _, p := range r.profiles {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, nil
}
func (r *memLaunchRepo) GetDefault(context.Context) (*launcher.LaunchProfile, error) {
	for _, p := range r.profiles {
		if p.IsDefault {
			return p, nil
		}
	}
	return nil, nil
}
func (r *memLaunchRepo) Save(context.Context, *launcher.LaunchProfile) error { return nil }
func (r *memLaunchRepo) Delete(context.Context, string) error                { return nil }

func TestSetLastLaunchProfileOnSuccess(t *testing.T) {
	repo := &memSettingsRepo{m: map[string]string{}}
	a := &App{ctx: context.Background(), settings: usecase.NewSettingsUseCase(repo)}
	if err := a.setLastLaunchProfileOnSuccess("p1", errors.New("launch failed")); err == nil {
		t.Fatal("expected launch error")
	}
	if repo.m["last_launch_profile_id"] != "" {
		t.Fatal("should not persist on launch failure")
	}
	if err := a.setLastLaunchProfileOnSuccess("p1", nil); err != nil {
		t.Fatalf("setLastLaunchProfileOnSuccess: %v", err)
	}
	if repo.m["last_launch_profile_id"] != "p1" {
		t.Fatalf("stored=%q", repo.m["last_launch_profile_id"])
	}
}
