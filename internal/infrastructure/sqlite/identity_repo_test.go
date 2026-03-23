package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"vrchat-tweaker/internal/domain/identity"
)

// mergeFromLogSave mirrors ActivityUseCase log cache updates: load, merge in domain, save.
func mergeFromLogSave(ctx context.Context, repo *UserCacheRepository, vrcID, displayName string, at time.Time) error {
	u, err := repo.GetByVRCUserID(ctx, vrcID)
	if err != nil {
		return err
	}
	if u == nil {
		u = &identity.UserCache{VRCUserID: vrcID, UserKind: identity.UserKindContact}
	}
	u.MergeFromLog(displayName, at)
	return repo.Save(ctx, u)
}

func TestUserCacheRepository_logMerge_lastContactNoRegression(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, dbErr := sql.Open("sqlite", dbPath)
	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer func() { _ = db.Close() }()
	if migrateErr := applySchema(db); migrateErr != nil {
		t.Fatal(migrateErr)
	}

	repo := NewUserCacheRepository(db)
	ctx := context.Background()
	const vrcID = "usr_lc_regress_test"

	t1 := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)

	err := mergeFromLogSave(ctx, repo, vrcID, "UserA", t1)
	if err != nil {
		t.Fatal(err)
	}
	u, err := repo.GetByVRCUserID(ctx, vrcID)
	if err != nil {
		t.Fatal(err)
	}
	if u.LastContactAt == nil || !u.LastContactAt.Equal(t1) {
		t.Fatalf("after first save LastContactAt = %v, want %v", u.LastContactAt, t1)
	}

	err = mergeFromLogSave(ctx, repo, vrcID, "UserA", t3)
	if err != nil {
		t.Fatal(err)
	}
	u, err = repo.GetByVRCUserID(ctx, vrcID)
	if err != nil {
		t.Fatal(err)
	}
	if u.LastContactAt == nil || !u.LastContactAt.Equal(t3) {
		t.Fatalf("after newer save LastContactAt = %v, want %v", u.LastContactAt, t3)
	}

	err = mergeFromLogSave(ctx, repo, vrcID, "UserA", t2)
	if err != nil {
		t.Fatal(err)
	}
	u, err = repo.GetByVRCUserID(ctx, vrcID)
	if err != nil {
		t.Fatal(err)
	}
	if u.LastContactAt == nil || !u.LastContactAt.Equal(t3) {
		t.Fatalf("after older save LastContactAt = %v, want %v (no regression)", u.LastContactAt, t3)
	}
	if u.UserKind != identity.UserKindContact {
		t.Fatalf("user_kind = %q, want contact", u.UserKind)
	}
}

func TestUserCacheRepository_logMerge_preservesFriendKind(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, dbErr := sql.Open("sqlite", dbPath)
	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer func() { _ = db.Close() }()
	if migrateErr := applySchema(db); migrateErr != nil {
		t.Fatal(migrateErr)
	}

	repo := NewUserCacheRepository(db)
	ctx := context.Background()
	const vrcID = "usr_friend_keep_kind"
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	f := &identity.UserCache{
		VRCUserID:   vrcID,
		DisplayName: "Friend",
		Status:      "offline",
		UserKind:    identity.UserKindFriend,
		LastUpdated: now,
	}
	if err := repo.Save(ctx, f); err != nil {
		t.Fatal(err)
	}
	if err := mergeFromLogSave(ctx, repo, vrcID, "FriendFromLog", now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	u, err := repo.GetByVRCUserID(ctx, vrcID)
	if err != nil {
		t.Fatal(err)
	}
	if u.UserKind != identity.UserKindFriend {
		t.Fatalf("user_kind = %q, want friend after log merge+save", u.UserKind)
	}
}

func TestUserCacheRepository_UpsertSelf_promotesExistingContactRow(t *testing.T) {
	dir := t.TempDir()
	db, dbErr := sql.Open("sqlite", filepath.Join(dir, "t.db"))
	if dbErr != nil {
		t.Fatal(dbErr)
	}
	defer func() { _ = db.Close() }()
	if migErr := applySchema(db); migErr != nil {
		t.Fatal(migErr)
	}
	repo := NewUserCacheRepository(db)
	ctx := context.Background()
	at := time.Date(2026, 3, 22, 12, 0, 0, 0, time.UTC)
	const vrcID = "usr_local_player"
	if err := mergeFromLogSave(ctx, repo, vrcID, "LocalUser", at); err != nil {
		t.Fatal(err)
	}
	before, err := repo.GetByVRCUserID(ctx, vrcID)
	if err != nil {
		t.Fatal(err)
	}
	if before.UserKind != identity.UserKindContact {
		t.Fatalf("want contact before UpsertSelf, got %q", before.UserKind)
	}
	self := &identity.UserCache{
		VRCUserID:                   vrcID,
		DisplayName:                 "DisplayFromAPI",
		Status:                      "active",
		UserKind:                    identity.UserKindSelf,
		LastUpdated:                 at.Add(time.Hour),
		SessionFingerprint:          "fp-local",
		Username:                    "localuser",
		StatusDescription:           "bio",
		UserState:                   "online",
		AvatarThumbnailURL:          "http://thumb",
		UserIconURL:                 "http://icon",
		ProfilePicOverrideThumbnail: "http://pic",
	}
	if upErr := repo.UpsertSelf(ctx, self); upErr != nil {
		t.Fatalf("UpsertSelf after contact row: %v", upErr)
	}
	after, err := repo.GetByVRCUserID(ctx, vrcID)
	if err != nil {
		t.Fatal(err)
	}
	if after.UserKind != identity.UserKindSelf {
		t.Fatalf("user_kind = %q, want self", after.UserKind)
	}
	if after.DisplayName != "DisplayFromAPI" || after.Username != "localuser" {
		t.Fatalf("profile not applied: %+v", after)
	}
	if after.LastContactAt == nil || !after.LastContactAt.Equal(at) {
		t.Fatalf("last_contact_at should be preserved from log, got %v want %v", after.LastContactAt, at)
	}
}

func TestUserCacheRepository_List_onlyFriendsWithStatus(t *testing.T) {
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if migErr := applySchema(db); migErr != nil {
		t.Fatal(migErr)
	}
	repo := NewUserCacheRepository(db)
	ctx := context.Background()
	at := time.Now().UTC()

	contact := &identity.UserCache{
		VRCUserID:     "usr_c",
		DisplayName:   "ContactOnly",
		UserKind:      identity.UserKindContact,
		LastUpdated:   at,
		FirstSeenAt:   &at,
		LastContactAt: &at,
	}
	if saveErr := repo.Save(ctx, contact); saveErr != nil {
		t.Fatal(saveErr)
	}
	_ = repo.UpsertSelf(ctx, &identity.UserCache{
		VRCUserID:                   "usr_self",
		DisplayName:                 "Self",
		Status:                      "active",
		UserKind:                    identity.UserKindSelf,
		LastUpdated:                 at,
		SessionFingerprint:          "abc",
		Username:                    "me",
		StatusDescription:           "",
		UserState:                   "online",
		AvatarThumbnailURL:          "",
		UserIconURL:                 "",
		ProfilePicOverrideThumbnail: "",
	})
	fr := &identity.UserCache{
		VRCUserID:   "usr_f",
		DisplayName: "Friend",
		Status:      "join me",
		UserKind:    identity.UserKindFriend,
		LastUpdated: at,
	}
	if saveErr := repo.Save(ctx, fr); saveErr != nil {
		t.Fatal(saveErr)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].VRCUserID != "usr_f" {
		t.Fatalf("List = %+v, want single friend usr_f", list)
	}
}

func TestUserCacheRepository_SaveBatch_afterMergeFromAPIFriend_keepsSelf(t *testing.T) {
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if migErr := applySchema(db); migErr != nil {
		t.Fatal(migErr)
	}
	repo := NewUserCacheRepository(db)
	ctx := context.Background()
	at := time.Now().UTC()
	self := &identity.UserCache{
		VRCUserID:                   "usr_me",
		DisplayName:                 "OriginalMe",
		Status:                      "busy",
		UserKind:                    identity.UserKindSelf,
		LastUpdated:                 at,
		SessionFingerprint:          "fp1",
		Username:                    "meuser",
		StatusDescription:           "d",
		UserState:                   "offline",
		AvatarThumbnailURL:          "http://a",
		UserIconURL:                 "http://i",
		ProfilePicOverrideThumbnail: "http://p",
	}
	if upErr := repo.UpsertSelf(ctx, self); upErr != nil {
		t.Fatal(upErr)
	}
	existing, err := repo.GetByVRCUserID(ctx, "usr_me")
	if err != nil {
		t.Fatal(err)
	}
	apiSnap := &identity.UserCache{
		VRCUserID:   "usr_me",
		DisplayName: "FromFriendsAPI",
		Status:      "join me",
		UserKind:    identity.UserKindFriend,
		LastUpdated: at.Add(time.Hour),
	}
	existing.MergeFromAPIFriend(apiSnap)
	if batchErr := repo.SaveBatch(ctx, []*identity.UserCache{existing}); batchErr != nil {
		t.Fatal(batchErr)
	}
	u, err := repo.GetByVRCUserID(ctx, "usr_me")
	if err != nil {
		t.Fatal(err)
	}
	if u.UserKind != identity.UserKindSelf {
		t.Fatalf("user_kind = %q, want self", u.UserKind)
	}
	if u.DisplayName != "OriginalMe" || u.Status != "busy" {
		t.Fatalf("self row was overwritten: %+v", u)
	}
	if u.SessionFingerprint != "fp1" {
		t.Fatalf("session fingerprint lost")
	}
}

func TestUserCacheRepository_GetSelfBySessionFingerprint(t *testing.T) {
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if migErr := applySchema(db); migErr != nil {
		t.Fatal(migErr)
	}
	repo := NewUserCacheRepository(db)
	ctx := context.Background()
	at := time.Now().UTC()
	const fp = "deadbeef"
	self := &identity.UserCache{
		VRCUserID:                   "usr_x",
		DisplayName:                 "X",
		Status:                      "active",
		UserKind:                    identity.UserKindSelf,
		LastUpdated:                 at,
		SessionFingerprint:          fp,
		Username:                    "u",
		StatusDescription:           "sd",
		UserState:                   "st",
		AvatarThumbnailURL:          "ta",
		UserIconURL:                 "ui",
		ProfilePicOverrideThumbnail: "tp",
	}
	if upErr := repo.UpsertSelf(ctx, self); upErr != nil {
		t.Fatal(upErr)
	}
	got, err := repo.GetSelfBySessionFingerprint(ctx, fp)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.VRCUserID != "usr_x" || got.Username != "u" {
		t.Fatalf("got %+v", got)
	}
	miss, err := repo.GetSelfBySessionFingerprint(ctx, "wrong")
	if err != nil {
		t.Fatal(err)
	}
	if miss != nil {
		t.Fatalf("wrong fingerprint: want nil, got %+v", miss)
	}
}
