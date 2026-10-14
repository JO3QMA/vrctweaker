package appupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"vrchat-tweaker/internal/appversion"
)

// Downloader fetches release zips over HTTPS.
type Downloader struct {
	HTTPClient *http.Client
	UserAgent  string
}

func (d *Downloader) httpClient() *http.Client {
	if d != nil && d.HTTPClient != nil {
		return d.HTTPClient
	}
	return http.DefaultClient
}

func (d *Downloader) ua() string {
	if d != nil && strings.TrimSpace(d.UserAgent) != "" {
		return strings.TrimSpace(d.UserAgent)
	}
	return appversion.UserAgent()
}

// DownloadFile saves url to destPath and verifies expectedSHA256 (hex, lowercase).
func (d *Downloader) DownloadFile(ctx context.Context, url, destPath, expectedSHA256 string) error {
	expectedSHA256 = strings.ToLower(strings.TrimSpace(expectedSHA256))
	if len(expectedSHA256) != 64 {
		return fmt.Errorf("appupdate: invalid expected sha256")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", d.ua())

	resp, err := d.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("appupdate: download: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("appupdate: download HTTP %d", resp.StatusCode)
	}

	if mkdirErr := os.MkdirAll(filepath.Dir(destPath), 0o700); mkdirErr != nil {
		return fmt.Errorf("appupdate: mkdir: %w", mkdirErr)
	}
	tmp, err := os.CreateTemp(filepath.Dir(destPath), ".dl-*.partial")
	if err != nil {
		return fmt.Errorf("appupdate: temp file: %w", err)
	}
	tmpPath := tmp.Name()
	hasher := sha256.New()
	mw := io.MultiWriter(tmp, hasher)
	if _, copyErr := io.Copy(mw, resp.Body); copyErr != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("appupdate: write: %w", copyErr)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("appupdate: close: %w", err)
	}
	got := hex.EncodeToString(hasher.Sum(nil))
	if got != expectedSHA256 {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("appupdate: sha256 mismatch")
	}
	if err := os.Rename(tmpPath, destPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("appupdate: rename: %w", err)
	}
	return nil
}
