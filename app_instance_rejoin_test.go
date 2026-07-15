package main

import (
	"context"
	"errors"
	"strings"
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
func (errRejoinPlayRepo) Save(context.Context, *activity.PlaySession) error {
	return errors.New("db down")
}
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

type errListLaunchRepo struct {
	memLaunchRepo
}

func (*errListLaunchRepo) List(context.Context) ([]*launcher.LaunchProfile, error) {
	return nil, errors.New("db down")
}

func TestGetDashboardLaunchBlock_listProfilesError(t *testing.T) {
	a := &App{ctx: context.Background()}
	a.activity = usecase.NewActivityUseCase(&memRejoinPlayRepo{}, nil, &memSettingsRepo{m: map[string]string{}}, nil, nil)
	a.launcher = usecase.NewLauncherUseCase(&errListLaunchRepo{})
	a.settings = usecase.NewSettingsUseCase(&memSettingsRepo{m: map[string]string{}})
	got, err := a.GetDashboardLaunchBlock()
	if err == nil || got != nil {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestGetDashboardLaunchBlock_continuesWithoutRejoinOnRejoinInfraFailure(t *testing.T) {
	a := &App{ctx: context.Background()}
	a.activity = usecase.NewActivityUseCase(errRejoinPlayRepo{}, nil, &memSettingsRepo{m: map[string]string{}}, nil, nil)
	a.launcher = usecase.NewLauncherUseCase(&memLaunchRepo{profiles: []*launcher.LaunchProfile{{ID: "p1", Name: "P"}}})
	a.settings = usecase.NewSettingsUseCase(&memSettingsRepo{m: map[string]string{}})
	got, err := a.GetDashboardLaunchBlock()
	if err != nil || got == nil {
		t.Fatalf("got=%v err=%v", got, err)
	}
	if got.Rejoin != nil || got.SelectedProfileID != "p1" || len(got.Profiles) != 1 {
		t.Fatalf("got=%+v", got)
	}
}

func TestInstanceRejoin_emptyProfileID(t *testing.T) {
	a := &App{ctx: context.Background()}
	err := a.InstanceRejoin("", "s1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInstanceRejoin_emptyPlaySessionID(t *testing.T) {
	a := &App{ctx: context.Background()}
	err := a.InstanceRejoin("p1", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetDashboardLaunchBlock_noProfiles(t *testing.T) {
	a := &App{ctx: context.Background()}
	a.activity = usecase.NewActivityUseCase(&memRejoinPlayRepo{latest: &activity.PlaySession{
		ID: "s1", StartTime: time.Now().UTC(), InstanceID: "wrld_x:1~public",
	}}, nil, &memSettingsRepo{m: map[string]string{}}, nil, nil)
	a.launcher = usecase.NewLauncherUseCase(&memLaunchRepo{})
	a.settings = usecase.NewSettingsUseCase(&memSettingsRepo{m: map[string]string{}})
	got, err := a.GetDashboardLaunchBlock()
	if err != nil || got == nil || len(got.Profiles) != 0 || got.SelectedProfileID != "" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if got.Rejoin == nil || got.Rejoin.PlaySessionID != "s1" {
		t.Fatalf("rejoin=%+v", got.Rejoin)
	}
}

func TestGetDashboardLaunchBlock_withoutRejoin(t *testing.T) {
	a := &App{ctx: context.Background()}
	a.activity = usecase.NewActivityUseCase(&memRejoinPlayRepo{}, nil, &memSettingsRepo{m: map[string]string{}}, nil, nil)
	a.launcher = usecase.NewLauncherUseCase(&memLaunchRepo{profiles: []*launcher.LaunchProfile{{ID: "p1", Name: "P"}}})
	a.settings = usecase.NewSettingsUseCase(&memSettingsRepo{m: map[string]string{}})
	got, err := a.GetDashboardLaunchBlock()
	if err != nil || got == nil || got.Rejoin != nil || got.SelectedProfileID != "p1" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestGetDashboardLaunchBlock_withRejoin(t *testing.T) {
	inst := "wrld_test1111-1111-4111-8111-111111111101:42~public"
	wid := activity.WorldIDFromInstanceKey(inst)
	s := &activity.PlaySession{ID: "s1", StartTime: time.Now().UTC(), InstanceID: inst}
	a := &App{ctx: context.Background()}
	a.activity = usecase.NewActivityUseCase(&memRejoinPlayRepo{latest: s, byID: map[string]*activity.PlaySession{"s1": s}},
		nil, &memSettingsRepo{m: map[string]string{}}, nil, &memRejoinWorldRepo{names: map[string]string{wid: "Test World"}})
	a.launcher = usecase.NewLauncherUseCase(&memLaunchRepo{profiles: []*launcher.LaunchProfile{{ID: "p1", Name: "P"}}})
	a.settings = usecase.NewSettingsUseCase(&memSettingsRepo{m: map[string]string{}})
	got, err := a.GetDashboardLaunchBlock()
	if err != nil || got == nil || got.Rejoin == nil {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if got.Rejoin.WorldDisplayName != "Test World" || got.SelectedProfileID != "p1" || got.Rejoin.PlaySessionID != "s1" {
		t.Fatalf("got=%+v", got)
	}
}

func TestGetDashboardLaunchBlock_selectedLast(t *testing.T) {
	a := &App{ctx: context.Background()}
	a.activity = usecase.NewActivityUseCase(&memRejoinPlayRepo{}, nil, &memSettingsRepo{m: map[string]string{}}, nil, nil)
	a.launcher = usecase.NewLauncherUseCase(&memLaunchRepo{profiles: []*launcher.LaunchProfile{
		{ID: "p1", Name: "P1"},
		{ID: "p2", Name: "P2", IsDefault: true},
	}})
	a.settings = usecase.NewSettingsUseCase(&memSettingsRepo{m: map[string]string{
		"last_launch_profile_id": "p1",
	}})
	got, err := a.GetDashboardLaunchBlock()
	if err != nil || got == nil || got.SelectedProfileID != "p1" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestInstanceRejoin_profileNotFound(t *testing.T) {
	s := &activity.PlaySession{ID: "s1", StartTime: time.Now().UTC(), InstanceID: "wrld_x:1~public"}
	a := &App{ctx: context.Background()}
	a.activity = usecase.NewActivityUseCase(&memRejoinPlayRepo{latest: s, byID: map[string]*activity.PlaySession{"s1": s}},
		nil, &memSettingsRepo{m: map[string]string{}}, nil, nil)
	a.launcher = usecase.NewLauncherUseCase(&memLaunchRepo{profiles: []*launcher.LaunchProfile{{ID: "p1", Name: "P"}}})
	a.settings = usecase.NewSettingsUseCase(&memSettingsRepo{m: map[string]string{}})
	err := a.InstanceRejoin("missing", "s1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInstanceRejoin_stalePlaySessionID(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	old := &activity.PlaySession{ID: "old", StartTime: t0, InstanceID: "wrld_old:1~public"}
	newest := &activity.PlaySession{ID: "new", StartTime: t1, InstanceID: "wrld_new:2~public"}
	a := &App{ctx: context.Background()}
	a.activity = usecase.NewActivityUseCase(&memRejoinPlayRepo{
		latest: newest,
		byID:   map[string]*activity.PlaySession{"old": old, "new": newest},
	}, nil, &memSettingsRepo{m: map[string]string{}}, nil, nil)
	a.launcher = usecase.NewLauncherUseCase(&memLaunchRepo{profiles: []*launcher.LaunchProfile{{ID: "p1", Name: "P"}}})
	a.settings = usecase.NewSettingsUseCase(&memSettingsRepo{m: map[string]string{}})
	err := a.InstanceRejoin("p1", "old")
	if err == nil || !strings.Contains(err.Error(), "changed") {
		t.Fatalf("err=%v", err)
	}
}

type memRejoinPlayRepo struct {
	latest *activity.PlaySession
	byID   map[string]*activity.PlaySession
}

func (r *memRejoinPlayRepo) List(context.Context, time.Time, time.Time) ([]*activity.PlaySession, error) {
	return nil, nil
}
func (r *memRejoinPlayRepo) GetByID(_ context.Context, id string) (*activity.PlaySession, error) {
	if r.byID != nil {
		return r.byID[id], nil
	}
	if r.latest != nil && r.latest.ID == id {
		return r.latest, nil
	}
	return nil, nil
}
func (r *memRejoinPlayRepo) Save(context.Context, *activity.PlaySession) error { return nil }
func (r *memRejoinPlayRepo) FindLatestWithoutEndTime(context.Context) (*activity.PlaySession, error) {
	return nil, nil
}
func (r *memRejoinPlayRepo) FindLatestWithInstanceID(context.Context) (*activity.PlaySession, error) {
	return r.latest, nil
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

func (r *memLaunchRepo) List(context.Context) ([]*launcher.LaunchProfile, error) {
	return r.profiles, nil
}
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

func TestLaunchVRChat_successUpdatesLastLaunch(t *testing.T) {
	repo := &memSettingsRepo{m: map[string]string{}}
	profile := &launcher.LaunchProfile{ID: "p1", Name: "P", Arguments: ""}
	uc := usecase.NewLauncherUseCase(&memLaunchRepo{profiles: []*launcher.LaunchProfile{profile}})
	a := &App{
		ctx:      context.Background(),
		settings: usecase.NewSettingsUseCase(repo),
		launcher: uc,
	}
	// ponytail: LaunchVRChat needs a real steam path or it fails on cmd.Start in CI;
	// test Last update via setLastLaunchProfileOnSuccess path used by LaunchVRChat.
	if err := a.setLastLaunchProfileOnSuccess("p1", nil); err != nil {
		t.Fatalf("setLastLaunchProfileOnSuccess: %v", err)
	}
	if repo.m["last_launch_profile_id"] != "p1" {
		t.Fatalf("stored=%q", repo.m["last_launch_profile_id"])
	}
}

func TestLaunchVRChat_failureNoLastUpdate(t *testing.T) {
	repo := &memSettingsRepo{m: map[string]string{}}
	a := &App{ctx: context.Background(), settings: usecase.NewSettingsUseCase(repo)}
	if err := a.setLastLaunchProfileOnSuccess("p1", errors.New("start failed")); err == nil {
		t.Fatal("expected error")
	}
	if repo.m["last_launch_profile_id"] != "" {
		t.Fatal("should not update last on failure")
	}
}
