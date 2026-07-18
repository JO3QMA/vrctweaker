package usecase

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newCookieUC(t *testing.T) (*CookieLinkageUseCase, *SettingsUseCase, string) {
	t.Helper()
	repo := &fakeAppSettingsRepo{m: map[string]string{}}
	settings := NewSettingsUseCase(repo)
	dir := t.TempDir()
	uc := NewCookieLinkageUseCase(settings)
	uc.ConfigDirOverride = dir
	return uc, settings, dir
}

func ackCookie(t *testing.T, uc *CookieLinkageUseCase) {
	t.Helper()
	if err := uc.AcknowledgeRisk(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestGet_noFile(t *testing.T) {
	uc, _, _ := newCookieUC(t)
	st, err := uc.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if st.Enabled || st.SourceKind != CookieSourceNone {
		t.Fatalf("status %+v", st)
	}
}

func TestGet_noManagedLines(t *testing.T) {
	uc, _, dir := newCookieUC(t)
	path := filepath.Join(dir, "config")
	content := "--sleep-requests 1\n# comment\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err := uc.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.Enabled {
		t.Fatal("expected disabled")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != content {
		t.Fatalf("other lines changed: %q", got)
	}
}

func TestResolve_configTxtOnly(t *testing.T) {
	uc, _, dir := newCookieUC(t)
	txt := filepath.Join(dir, "config.txt")
	if err := os.WriteFile(txt, []byte("--cookies-from-browser chrome\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err := uc.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !st.Enabled || st.SourceKind != CookieSourceBrowser || st.Browser != "chrome" {
		t.Fatalf("status %+v", st)
	}
	if st.ConfigPath != txt {
		t.Fatalf("config path %q want %q", st.ConfigPath, txt)
	}
	ackCookie(t, uc)
	if setErr := uc.SetBrowserSource(context.Background(), "edge"); setErr != nil {
		t.Fatal(setErr)
	}
	got, err := os.ReadFile(txt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "--cookies-from-browser edge") {
		t.Fatalf("got %q", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "config")); !os.IsNotExist(err) {
		t.Fatal("should not create config when config.txt is the target")
	}
}

func TestResolve_prefersConfig(t *testing.T) {
	uc, _, dir := newCookieUC(t)
	cfg := filepath.Join(dir, "config")
	txt := filepath.Join(dir, "config.txt")
	if err := os.WriteFile(cfg, []byte("--cookies-from-browser chrome\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(txt, []byte("--cookies-from-browser firefox\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err := uc.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.Browser != "chrome" || st.ConfigPath != cfg {
		t.Fatalf("status %+v", st)
	}
	ackCookie(t, uc)
	if disErr := uc.Disable(context.Background()); disErr != nil {
		t.Fatal(disErr)
	}
	txtBody, err := os.ReadFile(txt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(txtBody), "firefox") {
		t.Fatalf("config.txt should be untouched: %q", txtBody)
	}
}

func TestUpsert_browserSource(t *testing.T) {
	uc, _, dir := newCookieUC(t)
	cfg := filepath.Join(dir, "config")
	if err := os.WriteFile(cfg, []byte("--sleep-requests 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ackCookie(t, uc)
	if err := uc.SetBrowserSource(context.Background(), "chrome"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	if !strings.Contains(s, "--sleep-requests 1") {
		t.Fatalf("lost other lines: %q", s)
	}
	if !strings.Contains(s, "--cookies-from-browser chrome") {
		t.Fatalf("missing managed: %q", s)
	}
	st, err := uc.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !st.Enabled || st.SourceKind != CookieSourceBrowser || st.Browser != "chrome" {
		t.Fatalf("status %+v", st)
	}
}

func TestUpsert_cookiesFileMissing(t *testing.T) {
	uc, _, _ := newCookieUC(t)
	ackCookie(t, uc)
	err := uc.SetCookiesFileSource(context.Background(), filepath.Join(t.TempDir(), "missing.txt"))
	if !errors.Is(err, ErrCookieLinkageCookiesFileMissing) {
		t.Fatalf("got %v", err)
	}
}

func TestUpsert_cookiesFileEmptyOK(t *testing.T) {
	uc, _, dir := newCookieUC(t)
	cookies := filepath.Join(dir, "cookies.txt")
	if err := os.WriteFile(cookies, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	ackCookie(t, uc)
	if err := uc.SetCookiesFileSource(context.Background(), cookies); err != nil {
		t.Fatal(err)
	}
	st, err := uc.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !st.Enabled || st.SourceKind != CookieSourceFile || st.CookiesFilePath != cookies {
		t.Fatalf("status %+v", st)
	}
}

func TestDisable_removesManagedOnly(t *testing.T) {
	uc, _, dir := newCookieUC(t)
	cfg := filepath.Join(dir, "config")
	if err := os.WriteFile(cfg, []byte("--cookies-from-browser chrome\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ackCookie(t, uc)
	if err := uc.Disable(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cfg); !os.IsNotExist(err) {
		t.Fatal("empty config should be removed")
	}

	if err := os.WriteFile(cfg, []byte("--sleep-requests 1\n--cookies-from-browser chrome\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := uc.Disable(context.Background()); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "--sleep-requests 1\n" && strings.TrimSpace(string(got)) != "--sleep-requests 1" {
		// allow trailing newline normalization
		if !strings.Contains(string(got), "--sleep-requests 1") || strings.Contains(string(got), "cookies") {
			t.Fatalf("got %q", got)
		}
	}
}

func TestUpsert_replacesUnsupported(t *testing.T) {
	uc, _, dir := newCookieUC(t)
	cfg := filepath.Join(dir, "config")
	body := "--cookies-from-browser firefox::youtube\n--cookies /tmp/x\n"
	if err := os.WriteFile(cfg, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err := uc.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !st.Enabled || st.SourceKind != CookieSourceUnsupported {
		t.Fatalf("status %+v", st)
	}
	ackCookie(t, uc)
	if setErr := uc.SetBrowserSource(context.Background(), "chrome"); setErr != nil {
		t.Fatal(setErr)
	}
	got, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	if strings.Count(s, "--cookies") != 1 || !strings.Contains(s, "--cookies-from-browser chrome") {
		t.Fatalf("got %q", s)
	}
	if strings.Contains(s, "firefox::") || strings.Contains(s, "/tmp/x") {
		t.Fatalf("old managed remained: %q", s)
	}
}

func TestGet_readFailure(t *testing.T) {
	uc, _, dir := newCookieUC(t)
	cfg := filepath.Join(dir, "config")
	if err := os.Mkdir(cfg, 0o755); err != nil {
		t.Fatal(err)
	}
	st, err := uc.GetStatus(context.Background())
	if err == nil {
		t.Fatalf("expected error, got status %+v", st)
	}
	if st.Enabled {
		t.Fatal("must not fake disabled on read failure")
	}
}

func TestWrite_keepsOriginalOnFailure(t *testing.T) {
	uc, _, dir := newCookieUC(t)
	cfg := filepath.Join(dir, "config")
	orig := "--cookies-from-browser chrome\n--sleep-requests 1\n"
	if err := os.WriteFile(cfg, []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}
	ackCookie(t, uc)
	prev := renameFile
	renameFile = func(string, string) error {
		return errors.New("rename failed")
	}
	t.Cleanup(func() { renameFile = prev })

	err := uc.SetBrowserSource(context.Background(), "edge")
	if err == nil {
		t.Fatal("expected rename error")
	}
	got, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != orig {
		t.Fatalf("original changed: %q", got)
	}
}

func TestWrite_requiresRiskAck(t *testing.T) {
	uc, _, _ := newCookieUC(t)
	err := uc.SetBrowserSource(context.Background(), "chrome")
	if !errors.Is(err, ErrCookieLinkageRiskAckRequired) {
		t.Fatalf("got %v", err)
	}
}

func TestUpsert_cookiesFileRejectsNewline(t *testing.T) {
	uc, _, dir := newCookieUC(t)
	// Create a real file so Stat would pass if we only checked existence;
	// the newline must be rejected before write.
	real := filepath.Join(dir, "cookies.txt")
	if err := os.WriteFile(real, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	ackCookie(t, uc)
	evil := real + "\n--exec=calc.exe"
	err := uc.SetCookiesFileSource(context.Background(), evil)
	if !errors.Is(err, ErrCookieLinkageCookiesFileMissing) {
		t.Fatalf("got %v", err)
	}
}

func TestUpsert_tabSeparatedManagedLine(t *testing.T) {
	uc, _, dir := newCookieUC(t)
	cfg := filepath.Join(dir, "config")
	if err := os.WriteFile(cfg, []byte("--cookies-from-browser\tfirefox\n--sleep-requests 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err := uc.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !st.Enabled || st.SourceKind != CookieSourceBrowser || st.Browser != "firefox" {
		t.Fatalf("status %+v", st)
	}
	ackCookie(t, uc)
	if setErr := uc.SetBrowserSource(context.Background(), "chrome"); setErr != nil {
		t.Fatal(setErr)
	}
	got, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	if strings.Contains(s, "firefox") || strings.Count(s, "cookies-from-browser") != 1 {
		t.Fatalf("got %q", s)
	}
	if !strings.Contains(s, "--sleep-requests 1") {
		t.Fatalf("lost other line: %q", s)
	}
}

func TestWrite_preservesFileMode(t *testing.T) {
	uc, _, dir := newCookieUC(t)
	cfg := filepath.Join(dir, "config")
	if err := os.WriteFile(cfg, []byte("--cookies-from-browser chrome\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ackCookie(t, uc)
	if err := uc.SetBrowserSource(context.Background(), "edge"); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Fatalf("mode %o want 0600", fi.Mode().Perm())
	}
}
