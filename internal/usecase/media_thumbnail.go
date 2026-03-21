package usecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // register PNG decoder
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
	"vrchat-tweaker/internal/domain/media"
)

const (
	screenshotThumbMaxEdge     = 400
	screenshotThumbJPEGQuality = 80
	screenshotMaxSourceBytes   = 40 << 20 // 40 MiB
)

var (
	errScreenshotNotFound = errors.New("screenshot not found")

	// ErrScreenshotThumbnailNotCached is returned by ScreenshotThumbnailDataURL when
	// no valid cached thumbnail exists (thumbnails are built at ingest time).
	ErrScreenshotThumbnailNotCached = errors.New("screenshot thumbnail not cached")
)

func isJpegBlob(b []byte) bool {
	return len(b) >= 3 && b[0] == 0xff && b[1] == 0xd8 && b[2] == 0xff
}

// EnsureScreenshotThumbnail decodes the screenshot file, builds a JPEG thumbnail,
// and upserts it into the repository when missing or stale (by source size+mtime).
func (uc *MediaUseCase) EnsureScreenshotThumbnail(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("screenshot id is empty")
	}
	s, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if s == nil {
		return errScreenshotNotFound
	}

	path := filepath.Clean(s.FilePath)
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg":
	default:
		return fmt.Errorf("unsupported image extension: %s", ext)
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat screenshot file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("screenshot path is not a regular file")
	}
	sourceSize := info.Size()
	if sourceSize > screenshotMaxSourceBytes {
		return fmt.Errorf("screenshot file too large")
	}
	sourceModUnix := info.ModTime().Unix()

	cached, err := uc.repo.GetThumbnail(ctx, id)
	if err != nil {
		return err
	}
	if cached != nil && cached.SourceSize == sourceSize && cached.SourceModUnix == sourceModUnix && isJpegBlob(cached.JpegBlob) {
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open screenshot: %w", err)
	}
	defer f.Close()

	limited := io.LimitReader(f, screenshotMaxSourceBytes+1)
	img, format, err := image.Decode(limited)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}
	if format != "png" && format != "jpeg" {
		return fmt.Errorf("unsupported decoded format: %s", format)
	}

	out := resizeScreenshotThumb(img)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, out, &jpeg.Options{Quality: screenshotThumbJPEGQuality}); err != nil {
		return fmt.Errorf("encode jpeg: %w", err)
	}
	jpegBytes := buf.Bytes()

	thumb := &media.ScreenshotThumbnail{
		JpegBlob:      jpegBytes,
		SourceSize:    sourceSize,
		SourceModUnix: sourceModUnix,
	}
	if err := uc.repo.UpsertThumbnail(ctx, id, thumb); err != nil {
		return err
	}
	return nil
}

// ScreenshotThumbnailDataURL returns a data URL (JPEG) only when a valid cached
// thumbnail exists for the current file size and mtime. It does not generate thumbnails.
func (uc *MediaUseCase) ScreenshotThumbnailDataURL(ctx context.Context, id string) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("screenshot id is empty")
	}
	s, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	if s == nil {
		return "", errScreenshotNotFound
	}

	path := filepath.Clean(s.FilePath)
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg":
	default:
		return "", fmt.Errorf("unsupported image extension: %s", ext)
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("stat screenshot file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("screenshot path is not a regular file")
	}
	sourceSize := info.Size()
	if sourceSize > screenshotMaxSourceBytes {
		return "", fmt.Errorf("screenshot file too large")
	}
	sourceModUnix := info.ModTime().Unix()

	cached, err := uc.repo.GetThumbnail(ctx, id)
	if err != nil {
		return "", err
	}
	if cached != nil && cached.SourceSize == sourceSize && cached.SourceModUnix == sourceModUnix && isJpegBlob(cached.JpegBlob) {
		enc := base64.StdEncoding.EncodeToString(cached.JpegBlob)
		return "data:image/jpeg;base64," + enc, nil
	}

	return "", ErrScreenshotThumbnailNotCached
}

func resizeScreenshotThumb(img image.Image) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return img
	}
	if w <= screenshotThumbMaxEdge && h <= screenshotThumbMaxEdge {
		return img
	}
	var nw, nh int
	if w >= h {
		nw = screenshotThumbMaxEdge
		nh = int((int64(h) * int64(screenshotThumbMaxEdge)) / int64(w))
		if nh < 1 {
			nh = 1
		}
	} else {
		nh = screenshotThumbMaxEdge
		nw = int((int64(w) * int64(screenshotThumbMaxEdge)) / int64(h))
		if nw < 1 {
			nw = 1
		}
	}
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
	return dst
}
