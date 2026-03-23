package usecase

import (
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"vrchat-tweaker/internal/domain/media"
)

func TestMediaUseCase_ScreenshotThumbnailDataURL_OK(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "shot.png")
	if err := writeTestPNG(imgPath, 80, 60); err != nil {
		t.Fatalf("writeTestPNG: %v", err)
	}

	repo := newMockScreenshotRepo()
	s := &media.Screenshot{
		ID:       "thumb-id-1",
		FilePath: imgPath,
	}
	if err := repo.Save(context.Background(), s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	uc := NewMediaUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	if err := uc.EnsureScreenshotThumbnail(ctx, "thumb-id-1"); err != nil {
		t.Fatalf("EnsureScreenshotThumbnail: %v", err)
	}
	dataURL, err := uc.ScreenshotThumbnailDataURL(ctx, "thumb-id-1")
	if err != nil {
		t.Fatalf("ScreenshotThumbnailDataURL: %v", err)
	}
	if !strings.HasPrefix(dataURL, "data:image/jpeg;base64,") {
		t.Fatalf("want data:image/jpeg;base64, prefix, got prefix %q", dataURL[:min(40, len(dataURL))])
	}
}

func TestMediaUseCase_ScreenshotThumbnailDataURL_CacheSecondCallNoUpsert(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "shot.png")
	if err := writeTestPNG(imgPath, 80, 60); err != nil {
		t.Fatalf("writeTestPNG: %v", err)
	}

	repo := newMockScreenshotRepo()
	s := &media.Screenshot{
		ID:       "thumb-cache-1",
		FilePath: imgPath,
	}
	if err := repo.Save(context.Background(), s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	uc := NewMediaUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	if err := uc.EnsureScreenshotThumbnail(ctx, "thumb-cache-1"); err != nil {
		t.Fatalf("EnsureScreenshotThumbnail: %v", err)
	}
	if repo.thumbUpsertCnt != 1 {
		t.Fatalf("after ensure thumbUpsertCnt = %d, want 1", repo.thumbUpsertCnt)
	}
	if _, err := uc.ScreenshotThumbnailDataURL(ctx, "thumb-cache-1"); err != nil {
		t.Fatalf("first ScreenshotThumbnailDataURL: %v", err)
	}
	if _, err := uc.ScreenshotThumbnailDataURL(ctx, "thumb-cache-1"); err != nil {
		t.Fatalf("second ScreenshotThumbnailDataURL: %v", err)
	}
	if repo.thumbUpsertCnt != 1 {
		t.Fatalf("after two DataURL calls thumbUpsertCnt = %d, want 1 (cache hit, no second upsert)", repo.thumbUpsertCnt)
	}
}

func TestMediaUseCase_ScreenshotThumbnailDataURL_EnsuresWhenNotCached(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "shot.png")
	if err := writeTestPNG(imgPath, 40, 30); err != nil {
		t.Fatalf("writeTestPNG: %v", err)
	}
	repo := newMockScreenshotRepo()
	s := &media.Screenshot{ID: "lazy-thumb", FilePath: imgPath}
	if err := repo.Save(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	uc := NewMediaUseCase(repo, nil, nil, nil)
	ctx := context.Background()
	dataURL, err := uc.ScreenshotThumbnailDataURL(ctx, "lazy-thumb")
	if err != nil {
		t.Fatalf("ScreenshotThumbnailDataURL: %v", err)
	}
	if !strings.HasPrefix(dataURL, "data:image/jpeg;base64,") {
		t.Fatalf("want jpeg data URL, got prefix %q", dataURL[:min(40, len(dataURL))])
	}
	if repo.thumbUpsertCnt != 1 {
		t.Fatalf("thumbUpsertCnt = %d, want 1 after lazy ensure", repo.thumbUpsertCnt)
	}
	if _, err := uc.ScreenshotThumbnailDataURL(ctx, "lazy-thumb"); err != nil {
		t.Fatalf("second ScreenshotThumbnailDataURL: %v", err)
	}
	if repo.thumbUpsertCnt != 1 {
		t.Fatalf("thumbUpsertCnt = %d, want 1 after second call (cached)", repo.thumbUpsertCnt)
	}
}

func TestMediaUseCase_ScreenshotThumbnailDataURL_NotFound(t *testing.T) {
	repo := newMockScreenshotRepo()
	uc := NewMediaUseCase(repo, nil, nil, nil)
	ctx := context.Background()

	_, err := uc.ScreenshotThumbnailDataURL(ctx, "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, errScreenshotNotFound) {
		t.Fatalf("want errScreenshotNotFound, got %v", err)
	}
}

func TestMediaUseCase_ScreenshotThumbnailDataURL_EmptyID(t *testing.T) {
	uc := NewMediaUseCase(newMockScreenshotRepo(), nil, nil, nil)
	_, err := uc.ScreenshotThumbnailDataURL(context.Background(), "  ")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMediaUseCase_ScreenshotThumbnailDataURL_UnsupportedExt(t *testing.T) {
	dir := t.TempDir()
	badPath := filepath.Join(dir, "x.gif")
	if err := os.WriteFile(badPath, []byte("GIF87a"), 0644); err != nil {
		t.Fatal(err)
	}
	repo := newMockScreenshotRepo()
	s := &media.Screenshot{ID: "g1", FilePath: badPath}
	_ = repo.Save(context.Background(), s)
	uc := NewMediaUseCase(repo, nil, nil, nil)
	_, err := uc.ScreenshotThumbnailDataURL(context.Background(), "g1")
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("want unsupported extension error, got %v", err)
	}
}

func writeTestPNG(path string, w, h int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(x * 3), G: uint8(y * 3), B: 100, A: 255})
		}
	}
	return png.Encode(f, img)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
