package usecase

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	worldID   string
	worldName string
}

func (m *mockMetadataExtractor) Extract(path string) (worldID, worldName string, takenAt *time.Time, err error) {
	return m.worldID, m.worldName, nil, nil
}

func TestMediaUseCase_ScanDirectory_WithExtractor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "VRChat_wrld_test123.png")
	_ = os.WriteFile(path, []byte("fake"), 0644)

	repo := newMockScreenshotRepo()
	extractor := &mockMetadataExtractor{worldID: "wrld_test123", worldName: "Test World"}
	uc := NewMediaUseCase(repo, extractor)
	ctx := context.Background()

	count, err := uc.ScanDirectory(ctx, dir)
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
	uc := NewMediaUseCase(repo, extractor)
	ctx := context.Background()

	count, err := uc.ScanDirectory(ctx, dir)
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
	uc := NewMediaUseCase(repo, extractor)
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
