package main

import (
	"context"
	"testing"

	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/infrastructure/vrchatapi"
	"vrchat-tweaker/internal/usecase"
)

func TestGetPresenceChangeSection_notLoggedIn(t *testing.T) {
	a := &App{ctx: context.Background()}
	a.presenceChange = usecase.NewPresenceChangeUseCase(
		newAppPresenceIdentity(t, false).uc,
		&memSettingsRepo{m: map[string]string{}},
	)
	got, err := a.GetPresenceChangeSection()
	if err != nil || got == nil || got.LoggedIn {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestGetPresenceChangeSection_loggedIn(t *testing.T) {
	a := &App{ctx: context.Background()}
	idUC := newAppPresenceIdentity(t, true)
	idUC.selfRow.Status = "busy"
	idUC.selfRow.StatusDescription = "work"
	idUC.userRepo.getSelfRow = idUC.selfRow
	a.presenceChange = usecase.NewPresenceChangeUseCase(idUC.uc, &memSettingsRepo{m: map[string]string{}})
	got, err := a.GetPresenceChangeSection()
	if err != nil || got == nil || !got.LoggedIn || got.Status != "busy" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestApplyPresenceChange_success(t *testing.T) {
	a := &App{ctx: context.Background()}
	idUC := newAppPresenceIdentity(t, true)
	a.presenceChange = usecase.NewPresenceChangeUseCase(idUC.uc, &memSettingsRepo{m: map[string]string{}})
	got, err := a.ApplyPresenceChange("ask me", "hello")
	if err != nil || got == nil || got.Status != "busy" {
		t.Fatalf("got=%+v err=%v api=%q", got, err, idUC.api.lastBothStatus)
	}
	if idUC.api.setBothCalls != 1 {
		t.Fatalf("setBothCalls=%d", idUC.api.setBothCalls)
	}
}

type appPresenceIdentity struct {
	uc       *usecase.IdentityUseCase
	api      *mockAPIClientPresence
	userRepo *mockUserCacheRepoPresence
	selfRow  *identity.UserCache
}

type mockAPIClientPresence struct {
	token          string
	getCurrent     *vrchatapi.CurrentUserProfile
	setBothCalls   int
	lastBothStatus vrchatapi.UserStatus
	lastBothDesc   string
}

func (m *mockAPIClientPresence) Login(context.Context, string, string, string) (string, error) {
	return "", nil
}
func (m *mockAPIClientPresence) SetAuthToken(token string) { m.token = token }
func (m *mockAPIClientPresence) GetAuthToken() string      { return m.token }
func (m *mockAPIClientPresence) GetCurrentUser(context.Context) (*vrchatapi.CurrentUserProfile, error) {
	return m.getCurrent, nil
}
func (m *mockAPIClientPresence) GetFriends(context.Context) ([]vrchatapi.Friend, error) {
	return nil, nil
}
func (m *mockAPIClientPresence) GetUser(context.Context, string) (*vrchatapi.Friend, error) {
	return nil, nil
}
func (m *mockAPIClientPresence) SetUserStatus(context.Context, string, vrchatapi.UserStatus) error {
	return nil
}
func (m *mockAPIClientPresence) SetUserStatusDescription(context.Context, string, string) error {
	return nil
}
func (m *mockAPIClientPresence) SetUserStatusAndDescription(_ context.Context, _ string, status vrchatapi.UserStatus, description string) error {
	m.setBothCalls++
	m.lastBothStatus = status
	m.lastBothDesc = description
	return nil
}

type mockUserCacheRepoPresence struct {
	getSelfRow *identity.UserCache
	getSelfErr error
}

func (m *mockUserCacheRepoPresence) List(context.Context) ([]*identity.UserCache, error) {
	return nil, nil
}
func (m *mockUserCacheRepoPresence) GetByVRCUserID(context.Context, string) (*identity.UserCache, error) {
	return nil, nil
}
func (m *mockUserCacheRepoPresence) ListFavorites(context.Context) ([]*identity.UserCache, error) {
	return nil, nil
}
func (m *mockUserCacheRepoPresence) Save(context.Context, *identity.UserCache) error { return nil }
func (m *mockUserCacheRepoPresence) SaveBatch(context.Context, []*identity.UserCache) error {
	return nil
}
func (m *mockUserCacheRepoPresence) Delete(context.Context, string) error     { return nil }
func (m *mockUserCacheRepoPresence) DeleteAll(context.Context) (int64, error) { return 0, nil }
func (m *mockUserCacheRepoPresence) GetSelfBySessionFingerprint(context.Context, string) (*identity.UserCache, error) {
	if m.getSelfErr != nil {
		return nil, m.getSelfErr
	}
	return m.getSelfRow, nil
}
func (m *mockUserCacheRepoPresence) UpsertSelf(_ context.Context, u *identity.UserCache) error {
	if u != nil {
		cpy := *u
		m.getSelfRow = &cpy
	}
	return nil
}
func (m *mockUserCacheRepoPresence) DeleteSelfRows(context.Context) error { return nil }
func (m *mockUserCacheRepoPresence) ListContactsNeedingProfileResolution(context.Context) ([]*identity.UserCache, error) {
	return nil, nil
}

func newAppPresenceIdentity(t *testing.T, loggedIn bool) *appPresenceIdentity {
	t.Helper()
	api := &mockAPIClientPresence{}
	userRepo := &mockUserCacheRepoPresence{}
	uc := usecase.NewIdentityUseCase(userRepo, api, vrchatapi.NewStubCredentialStore(), &memSettingsRepo{m: map[string]string{}}, nil)
	h := &appPresenceIdentity{uc: uc, api: api, userRepo: userRepo}
	if !loggedIn {
		return h
	}
	api.SetAuthToken("token")
	fp := identity.AuthTokenFingerprint("token")
	h.selfRow = &identity.UserCache{
		VRCUserID:          "usr_self",
		Status:             "active",
		UserKind:           identity.UserKindSelf,
		SessionFingerprint: fp,
	}
	userRepo.getSelfRow = h.selfRow
	api.getCurrent = &vrchatapi.CurrentUserProfile{
		ID:                "usr_self",
		Status:            "busy",
		StatusDescription: "focus",
	}
	return h
}
