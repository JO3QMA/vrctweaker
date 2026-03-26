package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"vrchat-tweaker/internal/domain/activity"
	"vrchat-tweaker/internal/domain/identity"
	"vrchat-tweaker/internal/domain/media"
)

// mockScreenshotRepo implements media.ScreenshotRepository for tests.
type mockScreenshotRepo struct {
	screenshots    map[string]*media.Screenshot
	byPath         map[string]*media.Screenshot
	thumbs         map[string]*media.ScreenshotThumbnail
	listFilter     *media.ScreenshotFilter
	thumbUpsertCnt int
}

func newMockScreenshotRepo() *mockScreenshotRepo {
	return &mockScreenshotRepo{
		screenshots: make(map[string]*media.Screenshot),
		byPath:      make(map[string]*media.Screenshot),
		thumbs:      make(map[string]*media.ScreenshotThumbnail),
	}
}

func (m *mockScreenshotRepo) GetThumbnail(_ context.Context, screenshotID string) (*media.ScreenshotThumbnail, error) {
	t := m.thumbs[screenshotID]
	if t == nil {
		return nil, nil
	}
	cp := *t
	cp.JpegBlob = append([]byte(nil), t.JpegBlob...)
	return &cp, nil
}

func (m *mockScreenshotRepo) UpsertThumbnail(_ context.Context, screenshotID string, thumb *media.ScreenshotThumbnail) error {
	if thumb == nil {
		return nil
	}
	m.thumbUpsertCnt++
	cp := *thumb
	cp.JpegBlob = append([]byte(nil), thumb.JpegBlob...)
	m.thumbs[screenshotID] = &cp
	return nil
}

func (m *mockScreenshotRepo) DeleteThumbnail(_ context.Context, screenshotID string) error {
	delete(m.thumbs, screenshotID)
	return nil
}

func (m *mockScreenshotRepo) List(ctx context.Context, filter *media.ScreenshotFilter) ([]*media.Screenshot, error) {
	m.listFilter = filter
	var result []*media.Screenshot
	for _, s := range m.screenshots {
		if filter != nil && filter.FilePathPrefix != "" {
			if !hasPathPrefix(s.FilePath, filter.FilePathPrefix) {
				continue
			}
		}
		result = append(result, s)
	}
	return result, nil
}

func hasPathPrefix(path, prefix string) bool {
	normPath := filepath.ToSlash(path)
	normPrefix := filepath.ToSlash(prefix)
	return normPath == normPrefix || strings.HasPrefix(normPath, normPrefix)
}

func (m *mockScreenshotRepo) GetByID(ctx context.Context, id string) (*media.Screenshot, error) {
	return m.screenshots[id], nil
}

func (m *mockScreenshotRepo) GetByFilePath(ctx context.Context, filePath string) (*media.Screenshot, error) {
	return m.byPath[filePath], nil
}

func (m *mockScreenshotRepo) Save(ctx context.Context, s *media.Screenshot) error {
	m.screenshots[s.ID] = s
	m.byPath[s.FilePath] = s
	return nil
}

func (m *mockScreenshotRepo) Delete(ctx context.Context, id string) error {
	if s, ok := m.screenshots[id]; ok {
		delete(m.byPath, s.FilePath)
		delete(m.screenshots, id)
	}
	delete(m.thumbs, id)
	return nil
}

func (m *mockScreenshotRepo) DeleteAll(ctx context.Context) (int64, error) {
	n := int64(len(m.screenshots))
	m.screenshots = make(map[string]*media.Screenshot)
	m.byPath = make(map[string]*media.Screenshot)
	m.thumbs = make(map[string]*media.ScreenshotThumbnail)
	return n, nil
}

// mockMetadataExtractor returns fixed values for testing.
type mockMetadataExtractor struct {
	worldID           string
	worldName         string
	authorVRCUserID   string
	authorDisplayName string
}

func (m *mockMetadataExtractor) Extract(path string) (media.ScreenshotMetadata, error) {
	return media.ScreenshotMetadata{
		WorldID:           m.worldID,
		WorldDisplayName:  m.worldName,
		AuthorVRCUserID:   m.authorVRCUserID,
		AuthorDisplayName: m.authorDisplayName,
	}, nil
}

// mockWorldInfoRepo records WorldInfoRepository calls for tests.
type mockWorldInfoRepo struct {
	upsertDisplayCalls []struct {
		worldID, displayName string
		at                   time.Time
	}
	upsertVisitCalls []struct {
		worldID string
		at      time.Time
	}
}

func (m *mockWorldInfoRepo) UpsertVisit(_ context.Context, worldID string, at time.Time) error {
	m.upsertVisitCalls = append(m.upsertVisitCalls, struct {
		worldID string
		at      time.Time
	}{worldID, at})
	return nil
}

func (m *mockWorldInfoRepo) UpsertDisplayName(_ context.Context, worldID, displayName string, at time.Time) error {
	m.upsertDisplayCalls = append(m.upsertDisplayCalls, struct {
		worldID, displayName string
		at                   time.Time
	}{worldID, displayName, at})
	return nil
}

func (m *mockWorldInfoRepo) GetByWorldID(_ context.Context, _ string) (*activity.WorldInfo, error) {
	return nil, nil
}

type screenshotAuthorUserCacheMock struct {
	saved []*identity.UserCache
}

func (m *screenshotAuthorUserCacheMock) List(_ context.Context) ([]*identity.UserCache, error) {
	return nil, nil
}

func (m *screenshotAuthorUserCacheMock) GetByVRCUserID(_ context.Context, vrcUserID string) (*identity.UserCache, error) {
	for _, u := range m.saved {
		if u.VRCUserID == vrcUserID {
			cp := *u
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *screenshotAuthorUserCacheMock) ListFavorites(_ context.Context) ([]*identity.UserCache, error) {
	return nil, nil
}

func (m *screenshotAuthorUserCacheMock) Save(_ context.Context, u *identity.UserCache) error {
	if u == nil {
		return nil
	}
	cp := *u
	m.saved = append(m.saved, &cp)
	return nil
}

func (m *screenshotAuthorUserCacheMock) SaveBatch(_ context.Context, _ []*identity.UserCache) error {
	return nil
}
func (m *screenshotAuthorUserCacheMock) Delete(_ context.Context, _ string) error { return nil }
func (m *screenshotAuthorUserCacheMock) DeleteAll(_ context.Context) (int64, error) {
	return 0, nil
}
func (m *screenshotAuthorUserCacheMock) GetSelfBySessionFingerprint(_ context.Context, _ string) (*identity.UserCache, error) {
	return nil, nil
}
func (m *screenshotAuthorUserCacheMock) UpsertSelf(_ context.Context, _ *identity.UserCache) error {
	return nil
}
func (m *screenshotAuthorUserCacheMock) DeleteSelfRows(_ context.Context) error { return nil }

func TestMediaUseCase_IngestScreenshotFile_UpsertsAuthorInUserCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ingest_author.png")
	if err := writeTestPNG(path, 16, 16); err != nil {
		t.Fatalf("writeTestPNG: %v", err)
	}
	repo := newMockScreenshotRepo()
	userRepo := &screenshotAuthorUserCacheMock{}
	ext := &mockMetadataExtractor{
		worldID:           "wrld_a",
		worldName:         "W",
		authorVRCUserID:   "usr_testauthor",
		authorDisplayName: "Test Author",
	}
	uc := NewMediaUseCase(repo, ext, nil, userRepo)
	ctx := context.Background()
	_, _, err := uc.IngestScreenshotFile(ctx, path)
	if err != nil {
		t.Fatalf("IngestScreenshotFile: %v", err)
	}
	if len(userRepo.saved) != 1 {
		t.Fatalf("saved users = %d, want 1", len(userRepo.saved))
	}
	u := userRepo.saved[0]
	if u.VRCUserID != "usr_testauthor" || u.DisplayName != "Test Author" {
		t.Errorf("user = %+v", u)
	}
	if u.UserKind != identity.UserKindContact {
		t.Errorf("UserKind = %q", u.UserKind)
	}
}

func TestMediaUseCase_ScanDirectory_WithExtractor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "VRChat_wrld_test123.png")
	_ = os.WriteFile(path, []byte("fake"), 0644)

	repo := newMockScreenshotRepo()
	extractor := &mockMetadataExtractor{worldID: "wrld_test123", worldName: "Test World"}
	uc := NewMediaUseCase(repo, extractor, nil, nil)
	ctx := context.Background()

	count, err := uc.ScanDirectory(ctx, dir, nil)
	if err != nil {
		t.Fatalf("ScanDirectory: %v", err)
	}
	if count != 1 {
		t.Errorf("ScanDirectory: count = %d, want 1", count)
	}
	got, _ := repo.GetByFilePath(ctx, path)
	if got == nil {
		t.Fatal("screenshot not saved")
	}
	if got.WorldID != "wrld_test123" {
		t.Errorf("WorldID = %q, want wrld_test123", got.WorldID)
	}
	if got.WorldName != "Test World" {
		t.Errorf("WorldName = %q, want Test World", got.WorldName)
	}
	if got.FileSizeBytes == nil || *got.FileSizeBytes != 4 {
		t.Errorf("FileSizeBytes = %v, want 4", got.FileSizeBytes)
	}
}

func TestMediaUseCase_ScanDirectory_ExtractorErrorContinues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plain.png")
	_ = os.WriteFile(path, []byte("fake"), 0644)

	repo := newMockScreenshotRepo()
	extractor := media.NewDefaultMetadataExtractor()
	uc := NewMediaUseCase(repo, extractor, nil, nil)
	ctx := context.Background()

	count, err := uc.ScanDirectory(ctx, dir, nil)
	if err != nil {
		t.Fatalf("ScanDirectory: %v", err)
	}
	if count != 1 {
		t.Errorf("ScanDirectory: count = %d, want 1", count)
	}
	got, _ := repo.GetByFilePath(ctx, path)
	if got == nil {
		t.Fatal("screenshot not saved")
	}
	if got.WorldID != "" || got.WorldName != "" {
		t.Errorf("extraction failure should yield empty metadata, got worldID=%q worldName=%q", got.WorldID, got.WorldName)
	}
}

func TestMediaUseCase_IngestScreenshotFile_NewThenSkip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ingest.png")
	if err := writeTestPNG(path, 32, 24); err != nil {
		t.Fatalf("writeTestPNG: %v", err)
	}
	repo := newMockScreenshotRepo()
	uc := NewMediaUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	s, created, err := uc.IngestScreenshotFile(ctx, path)
	if err != nil {
		t.Fatalf("IngestScreenshotFile: %v", err)
	}
	if !created || s == nil {
		t.Fatalf("want new row, got created=%v s=%v", created, s)
	}
	if repo.thumbUpsertCnt != 1 {
		t.Fatalf("thumbUpsertCnt = %d, want 1", repo.thumbUpsertCnt)
	}

	s2, created2, err2 := uc.IngestScreenshotFile(ctx, path)
	if err2 != nil {
		t.Fatalf("second IngestScreenshotFile: %v", err2)
	}
	if created2 || s2 == nil || s2.ID != s.ID {
		t.Fatalf("second ingest: want skip same id, got created=%v id=%v", created2, s2)
	}
	if repo.thumbUpsertCnt != 1 {
		t.Fatalf("after second ingest thumbUpsertCnt = %d, want 1", repo.thumbUpsertCnt)
	}
}

func TestMediaUseCase_IngestScreenshotFile_SkipsNonImage(t *testing.T) {
	dir := t.TempDir()
	txt := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(txt, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	repo := newMockScreenshotRepo()
	uc := NewMediaUseCase(repo, nil, nil, nil)
	s, created, err := uc.IngestScreenshotFile(context.Background(), txt)
	if err != nil {
		t.Fatalf("IngestScreenshotFile: %v", err)
	}
	if created || s != nil {
		t.Fatalf("want skip non-image, got s=%v created=%v", s, created)
	}
}

func TestMediaUseCase_IngestUnderPictureRootSince_onlyNewerMtime(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "old.png")
	newPath := filepath.Join(dir, "new.png")
	_ = os.WriteFile(oldPath, []byte("a"), 0o644)
	_ = os.WriteFile(newPath, []byte("b"), 0o644)
	oldM := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	newM := time.Date(2030, 6, 7, 8, 9, 10, 0, time.UTC)
	if err := os.Chtimes(oldPath, oldM, oldM); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(newPath, newM, newM); err != nil {
		t.Fatal(err)
	}

	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	repo := newMockScreenshotRepo()
	uc := NewMediaUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	count, err := uc.IngestUnderPictureRootSince(ctx, dir, since)
	if err != nil {
		t.Fatalf("IngestUnderPictureRootSince: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
	if repo.byPath[oldPath] != nil {
		t.Error("old.png should not be ingested")
	}
	if repo.byPath[newPath] == nil {
		t.Fatal("new.png should be ingested")
	}
}

func TestMediaUseCase_IngestUnderPictureRootSince_notADir(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "file.png")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	uc := NewMediaUseCase(newMockScreenshotRepo(), nil, nil, nil)
	count, err := uc.IngestUnderPictureRootSince(context.Background(), f, time.Time{})
	if err != nil {
		t.Fatalf("IngestUnderPictureRootSince: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}

func TestMediaUseCase_IngestUnderPictureRootSince_contextCancelDuringWalk(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "n.png")
	_ = os.WriteFile(p, []byte("x"), 0o644)
	future := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	_ = os.Chtimes(p, future, future)

	repo := newMockScreenshotRepo()
	uc := NewMediaUseCase(repo, nil, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := uc.IngestUnderPictureRootSince(ctx, dir, time.Time{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestMediaUseCase_ReindexScreenshots(t *testing.T) {
	dir := t.TempDir()
	basePath := filepath.Join(dir, "screenshots")
	_ = os.MkdirAll(basePath, 0755)
	path := filepath.Join(basePath, "shot.png")
	_ = os.WriteFile(path, []byte("fake"), 0644)

	repo := newMockScreenshotRepo()
	s := &media.Screenshot{
		ID:        "id-1",
		FilePath:  path,
		WorldID:   "",
		WorldName: "",
		TakenAt:   nil,
	}
	_ = repo.Save(context.Background(), s)

	extractor := &mockMetadataExtractor{worldID: "wrld_reindexed", worldName: "Reindexed World"}
	uc := NewMediaUseCase(repo, extractor, nil, nil)
	ctx := context.Background()

	updated, err := uc.ReindexScreenshots(ctx, basePath)
	if err != nil {
		t.Fatalf("ReindexScreenshots: %v", err)
	}
	if updated != 1 {
		t.Errorf("ReindexScreenshots: updated = %d, want 1", updated)
	}

	got, _ := repo.GetByID(ctx, "id-1")
	if got.WorldID != "wrld_reindexed" {
		t.Errorf("WorldID = %q, want wrld_reindexed", got.WorldID)
	}
	if got.WorldName != "Reindexed World" {
		t.Errorf("WorldName = %q, want Reindexed World", got.WorldName)
	}
}

func TestMediaUseCase_IngestScreenshotFile_UpsertsWorldInfo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ingest_world.png")
	if err := writeTestPNG(path, 32, 24); err != nil {
		t.Fatalf("writeTestPNG: %v", err)
	}
	repo := newMockScreenshotRepo()
	worldRepo := &mockWorldInfoRepo{}
	extractor := &mockMetadataExtractor{worldID: "wrld_ingest", worldName: "Ingest World"}
	uc := NewMediaUseCase(repo, extractor, worldRepo, nil)
	ctx := context.Background()

	s, created, err := uc.IngestScreenshotFile(ctx, path)
	if err != nil {
		t.Fatalf("IngestScreenshotFile: %v", err)
	}
	if !created || s == nil {
		t.Fatalf("want new row")
	}
	if len(worldRepo.upsertDisplayCalls) != 1 {
		t.Fatalf("UpsertDisplayName calls = %d, want 1", len(worldRepo.upsertDisplayCalls))
	}
	if len(worldRepo.upsertVisitCalls) != 0 {
		t.Fatalf("UpsertVisit calls = %d, want 0", len(worldRepo.upsertVisitCalls))
	}
	c := worldRepo.upsertDisplayCalls[0]
	if c.worldID != "wrld_ingest" || c.displayName != "Ingest World" {
		t.Errorf("UpsertDisplayName args = (%q, %q), want (wrld_ingest, Ingest World)", c.worldID, c.displayName)
	}
	if s.TakenAt != nil && !c.at.Equal(*s.TakenAt) {
		t.Errorf("UpsertDisplayName at = %v, want screenshot TakenAt %v", c.at, *s.TakenAt)
	}
}

func TestMediaUseCase_IngestScreenshotFile_UpsertsWorldVisitWhenNameEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ingest_visit.png")
	if err := writeTestPNG(path, 16, 16); err != nil {
		t.Fatalf("writeTestPNG: %v", err)
	}
	repo := newMockScreenshotRepo()
	worldRepo := &mockWorldInfoRepo{}
	extractor := &mockMetadataExtractor{worldID: "wrld_noidname", worldName: ""}
	uc := NewMediaUseCase(repo, extractor, worldRepo, nil)
	ctx := context.Background()

	_, created, err := uc.IngestScreenshotFile(ctx, path)
	if err != nil {
		t.Fatalf("IngestScreenshotFile: %v", err)
	}
	if !created {
		t.Fatal("want new row")
	}
	if len(worldRepo.upsertVisitCalls) != 1 {
		t.Fatalf("UpsertVisit calls = %d, want 1", len(worldRepo.upsertVisitCalls))
	}
	if len(worldRepo.upsertDisplayCalls) != 0 {
		t.Fatalf("UpsertDisplayName calls = %d, want 0", len(worldRepo.upsertDisplayCalls))
	}
	if worldRepo.upsertVisitCalls[0].worldID != "wrld_noidname" {
		t.Errorf("worldID = %q", worldRepo.upsertVisitCalls[0].worldID)
	}
}

func TestMediaUseCase_ReindexScreenshots_UpsertsWorldInfo(t *testing.T) {
	dir := t.TempDir()
	basePath := filepath.Join(dir, "screenshots")
	_ = os.MkdirAll(basePath, 0755)
	path := filepath.Join(basePath, "reindex_world.png")
	_ = os.WriteFile(path, []byte("fake"), 0644)

	repo := newMockScreenshotRepo()
	s := &media.Screenshot{
		ID:        "id-rw",
		FilePath:  path,
		WorldID:   "",
		WorldName: "",
		TakenAt:   nil,
	}
	_ = repo.Save(context.Background(), s)

	worldRepo := &mockWorldInfoRepo{}
	extractor := &mockMetadataExtractor{worldID: "wrld_reidx", worldName: "Reidx World"}
	uc := NewMediaUseCase(repo, extractor, worldRepo, nil)
	ctx := context.Background()

	updated, err := uc.ReindexScreenshots(ctx, basePath)
	if err != nil {
		t.Fatalf("ReindexScreenshots: %v", err)
	}
	if updated != 1 {
		t.Fatalf("updated = %d, want 1", updated)
	}
	if len(worldRepo.upsertDisplayCalls) != 1 {
		t.Fatalf("UpsertDisplayName calls = %d, want 1", len(worldRepo.upsertDisplayCalls))
	}
	c := worldRepo.upsertDisplayCalls[0]
	if c.worldID != "wrld_reidx" || c.displayName != "Reidx World" {
		t.Errorf("UpsertDisplayName = (%q, %q)", c.worldID, c.displayName)
	}
}

func TestMediaUseCase_ScanDirectory_ProgressCallbacks(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.png")
	p2 := filepath.Join(dir, "b.png")
	_ = os.WriteFile(p1, []byte("fake1"), 0644)
	_ = os.WriteFile(p2, []byte("fake2"), 0644)

	repo := newMockScreenshotRepo()
	uc := NewMediaUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	var got []ScanProgress
	count, err := uc.ScanDirectory(ctx, dir, func(p ScanProgress) {
		got = append(got, p)
	})
	if err != nil {
		t.Fatalf("ScanDirectory: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	// listing (final 2/0) + importing 0/2 + two per-file steps
	if len(got) != 4 {
		t.Fatalf("len(progress) = %d, want 4: %#v", len(got), got)
	}
	if got[0].Phase != "listing" || got[0].Current != 2 || got[0].Total != 0 {
		t.Errorf("got[0] = %#v, want listing 2/0", got[0])
	}
	if got[1].Phase != "importing" || got[1].Current != 0 || got[1].Total != 2 || got[1].Item != "" {
		t.Errorf("got[1] = %#v, want importing 0/2", got[1])
	}
	var importItems []string
	for i := 2; i < 4; i++ {
		if got[i].Phase != "importing" || got[i].Total != 2 {
			t.Errorf("got[%d] = %#v, want importing n/2", i, got[i])
		}
		if got[i].Item != "" {
			importItems = append(importItems, got[i].Item)
		}
	}
	if got[2].Current != 1 || got[3].Current != 2 {
		t.Errorf("want Current 1 then 2, got %#v and %#v", got[2], got[3])
	}
	slices.Sort(importItems)
	if !slices.Equal(importItems, []string{"a.png", "b.png"}) {
		t.Errorf("import basenames = %v, want [a.png b.png]", importItems)
	}
}

func TestMediaUseCase_ScanDirectory_ContextCancelDuringWalk(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.png")
	_ = os.WriteFile(p1, []byte("fake1"), 0o644)

	repo := newMockScreenshotRepo()
	uc := NewMediaUseCase(repo, nil, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := uc.ScanDirectory(ctx, dir, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ScanDirectory: err = %v, want context.Canceled", err)
	}
}

func TestMediaUseCase_ScanDirectory_ContextCancelDuringImport(t *testing.T) {
	dir := t.TempDir()
	for i := range 5 {
		p := filepath.Join(dir, fmt.Sprintf("f%d.png", i))
		_ = os.WriteFile(p, []byte("x"), 0o644)
	}

	repo := newMockScreenshotRepo()
	uc := NewMediaUseCase(repo, nil, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())

	_, err := uc.ScanDirectory(ctx, dir, func(p ScanProgress) {
		if p.Phase == ScanPhaseImporting && p.Current >= 2 {
			cancel()
		}
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ScanDirectory: err = %v, want context.Canceled", err)
	}
}
