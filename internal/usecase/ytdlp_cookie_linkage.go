package usecase

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"unicode"
)

// Cookie source kinds for Cookie linkage effective state.
const (
	CookieSourceNone        = ""
	CookieSourceBrowser     = "browser"
	CookieSourceFile        = "file"
	CookieSourceUnsupported = "unsupported"
)

var (
	// ErrCookieLinkageRiskAckRequired is returned when writing without risk acknowledgment.
	ErrCookieLinkageRiskAckRequired = errors.New("cookie linkage risk acknowledgment required")
	// ErrCookieLinkageUnsupportedPlatform is returned on non-Windows (unless configDirOverride is set).
	ErrCookieLinkageUnsupportedPlatform = errors.New("yt-dlp Cookie linkage is Windows only")
	// ErrCookieLinkageCookiesFileMissing is returned when the cookies file path does not exist.
	ErrCookieLinkageCookiesFileMissing = errors.New("cookies file does not exist")
	// ErrCookieLinkageInvalidBrowser is returned for browsers outside the v1 allow-list.
	ErrCookieLinkageInvalidBrowser = errors.New("unsupported browser for cookie linkage")

	// renameFile is os.Rename; tests may stub it.
	renameFile = os.Rename
)

var cookieBrowsersV1 = map[string]struct{}{
	"chrome":  {},
	"edge":    {},
	"firefox": {},
}

// CookieLinkageStatus is the effective (+ ack) state for the Video tab UI.
type CookieLinkageStatus struct {
	Supported         bool   `json:"supported"`
	UnsupportedReason string `json:"unsupportedReason,omitempty"`
	Enabled           bool   `json:"enabled"`
	SourceKind        string `json:"sourceKind"`
	Browser           string `json:"browser,omitempty"`
	CookiesFilePath   string `json:"cookiesFilePath,omitempty"`
	ConfigPath        string `json:"configPath,omitempty"`
	RiskAcknowledged  bool   `json:"riskAcknowledged"`
}

// CookieLinkageUseCase manages Managed cookie options in yt-dlp user config.
type CookieLinkageUseCase struct {
	mu       sync.Mutex
	settings *SettingsUseCase
	// configDirOverride replaces %APPDATA%/yt-dlp (tests only). When set, skips Windows-only gate.
	configDirOverride string
}

// NewCookieLinkageUseCase wires settings for risk acknowledgment persistence.
func NewCookieLinkageUseCase(settings *SettingsUseCase) *CookieLinkageUseCase {
	if settings == nil {
		panic("usecase: NewCookieLinkageUseCase: settings is nil")
	}
	return &CookieLinkageUseCase{settings: settings}
}

// AcknowledgeRisk records Cookie linkage risk acknowledgment.
func (uc *CookieLinkageUseCase) AcknowledgeRisk(ctx context.Context) error {
	return uc.settings.SetYTDLPCookieLinkageRiskAck(ctx, true)
}

// GetStatus reads Cookie linkage effective state from yt-dlp user config.
func (uc *CookieLinkageUseCase) GetStatus(ctx context.Context) (CookieLinkageStatus, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return uc.getStatusLocked(ctx)
}

// SetBrowserSource upserts --cookies-from-browser <browser> (v1: chrome/edge/firefox).
func (uc *CookieLinkageUseCase) SetBrowserSource(ctx context.Context, browser string) error {
	browser = strings.ToLower(strings.TrimSpace(browser))
	if _, ok := cookieBrowsersV1[browser]; !ok {
		return ErrCookieLinkageInvalidBrowser
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if err := uc.requireWriteReady(ctx); err != nil {
		return err
	}
	line := "--cookies-from-browser " + browser
	return uc.writeManagedLocked(line)
}

// SetCookiesFileSource upserts --cookies <path>; path must exist (empty file OK).
func (uc *CookieLinkageUseCase) SetCookiesFileSource(ctx context.Context, path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return ErrCookieLinkageCookiesFileMissing
	}
	if strings.ContainsAny(path, "\n\r\"'") {
		return ErrCookieLinkageCookiesFileMissing
	}
	fi, err := os.Stat(path)
	if err != nil || fi.IsDir() {
		return ErrCookieLinkageCookiesFileMissing
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if err := uc.requireWriteReady(ctx); err != nil {
		return err
	}
	line := "--cookies " + quoteConfigArg(path)
	return uc.writeManagedLocked(line)
}

// Disable removes Managed cookie options only.
func (uc *CookieLinkageUseCase) Disable(ctx context.Context) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if err := uc.requireWriteReady(ctx); err != nil {
		return err
	}
	return uc.writeManagedLocked("")
}

func (uc *CookieLinkageUseCase) requireWriteReady(ctx context.Context) error {
	if !uc.platformOK() {
		return ErrCookieLinkageUnsupportedPlatform
	}
	ack, err := uc.settings.GetYTDLPCookieLinkageRiskAck(ctx)
	if err != nil {
		return err
	}
	if !ack {
		return ErrCookieLinkageRiskAckRequired
	}
	return nil
}

func (uc *CookieLinkageUseCase) platformOK() bool {
	if uc.configDirOverride != "" {
		return true
	}
	return runtime.GOOS == "windows"
}

func (uc *CookieLinkageUseCase) getStatusLocked(ctx context.Context) (CookieLinkageStatus, error) {
	st := CookieLinkageStatus{SourceKind: CookieSourceNone}
	if !uc.platformOK() {
		st.UnsupportedReason = "unsupportedPlatform"
		return st, nil
	}
	st.Supported = true
	ack, err := uc.settings.GetYTDLPCookieLinkageRiskAck(ctx)
	if err != nil {
		return st, err
	}
	st.RiskAcknowledged = ack

	path, err := uc.resolveConfigPathLocked()
	if err != nil {
		return CookieLinkageStatus{}, err
	}
	st.ConfigPath = path
	if path == "" {
		return st, nil
	}

	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			st.ConfigPath = ""
			return st, nil
		}
		return CookieLinkageStatus{}, fmt.Errorf("cookie linkage config read: %w", err)
	}
	if fi.IsDir() {
		return CookieLinkageStatus{}, fmt.Errorf("cookie linkage config read: path is a directory")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return CookieLinkageStatus{}, fmt.Errorf("cookie linkage config read: %w", err)
	}

	enabled, kind, browser, cookiesPath, err := parseManagedCookieOptions(string(raw))
	if err != nil {
		return CookieLinkageStatus{}, fmt.Errorf("cookie linkage config read: %w", err)
	}
	st.Enabled = enabled
	st.SourceKind = kind
	st.Browser = browser
	st.CookiesFilePath = cookiesPath
	return st, nil
}

// resolveConfigPathLocked returns the existing target path, or "" if neither file exists
// (caller will create config on write).
//
// yt-dlp User config loads the first existing candidate only (options.py:
// next(filter(None, _load_from_config_dirs(...)))). Within %APPDATA%/yt-dlp the
// order is config then config.txt — matching this preference when both exist.
func (uc *CookieLinkageUseCase) resolveConfigPathLocked() (string, error) {
	dir, err := uc.configDirLocked()
	if err != nil {
		return "", err
	}
	cfg := filepath.Join(dir, "config")
	txt := filepath.Join(dir, "config.txt")

	cfgInfo, cfgErr := os.Stat(cfg)
	if cfgErr == nil {
		if cfgInfo.IsDir() {
			return "", fmt.Errorf("cookie linkage config read: path is a directory")
		}
		return cfg, nil
	}
	if !os.IsNotExist(cfgErr) {
		return "", fmt.Errorf("cookie linkage config read: %w", cfgErr)
	}

	txtInfo, txtErr := os.Stat(txt)
	if txtErr == nil {
		if txtInfo.IsDir() {
			return "", fmt.Errorf("cookie linkage config read: path is a directory")
		}
		return txt, nil
	}
	if os.IsNotExist(txtErr) {
		return "", nil
	}
	return "", fmt.Errorf("cookie linkage config read: %w", txtErr)
}

func (uc *CookieLinkageUseCase) writeTargetPathLocked() (string, error) {
	path, err := uc.resolveConfigPathLocked()
	if err != nil {
		return "", err
	}
	if path != "" {
		return path, nil
	}
	dir, err := uc.configDirLocked()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config"), nil
}

func (uc *CookieLinkageUseCase) configDirLocked() (string, error) {
	if uc.configDirOverride != "" {
		return filepath.Clean(uc.configDirOverride), nil
	}
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", errors.New("APPDATA is not set")
	}
	return filepath.Join(appData, "yt-dlp"), nil
}

func (uc *CookieLinkageUseCase) writeManagedLocked(managedLine string) error {
	target, err := uc.writeTargetPathLocked()
	if err != nil {
		return err
	}
	// Read failure must block writes (do not treat as empty).
	var existing string
	var existingMode os.FileMode
	fi, statErr := os.Stat(target)
	switch {
	case statErr == nil && fi.IsDir():
		return fmt.Errorf("cookie linkage config read: path is a directory")
	case statErr == nil:
		raw, readErr := os.ReadFile(target)
		if readErr != nil {
			return fmt.Errorf("cookie linkage config read: %w", readErr)
		}
		existing = string(raw)
		existingMode = fi.Mode().Perm()
	case os.IsNotExist(statErr):
		existing = ""
	default:
		return fmt.Errorf("cookie linkage config read: %w", statErr)
	}

	next, err := upsertManagedCookieLines(existing, managedLine)
	if err != nil {
		return fmt.Errorf("cookie linkage config update: %w", err)
	}
	if strings.TrimSpace(next) == "" {
		if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove empty yt-dlp user config: %w", err)
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create yt-dlp config dir: %w", err)
	}
	return atomicWriteFile(target, []byte(next), existingMode)
}

func atomicWriteFile(path string, data []byte, existingMode os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".ytdlp-config-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp config: %w", err)
	}
	if existingMode != 0 {
		if err := tmp.Chmod(existingMode); err != nil {
			_ = tmp.Close()
			return fmt.Errorf("chmod temp config: %w", err)
		}
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp config: %w", err)
	}
	if err := renameFile(tmpName, path); err != nil {
		return fmt.Errorf("replace yt-dlp user config: %w", err)
	}
	return nil
}

func parseManagedCookieOptions(content string) (enabled bool, kind, browser, cookiesPath string, err error) {
	var browsers []string
	var files []string
	var unsupported bool

	sc := bufio.NewScanner(strings.NewReader(content))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if opt, rest, ok := splitConfigOption(line); ok {
			switch opt {
			case "--cookies-from-browser":
				enabled = true
				arg := strings.TrimSpace(rest)
				if arg == "" || strings.Contains(arg, ":") {
					unsupported = true
					continue
				}
				b := strings.ToLower(strings.Fields(arg)[0])
				if _, ok := cookieBrowsersV1[b]; !ok {
					unsupported = true
					continue
				}
				browsers = append(browsers, b)
			case "--cookies":
				enabled = true
				arg := strings.TrimSpace(rest)
				if arg == "" {
					unsupported = true
					continue
				}
				files = append(files, arg)
			}
		}
	}
	if err := sc.Err(); err != nil {
		return false, "", "", "", err
	}

	if !enabled {
		return false, CookieSourceNone, "", "", nil
	}
	if unsupported || (len(browsers) > 0 && len(files) > 0) || len(browsers) > 1 || len(files) > 1 {
		return true, CookieSourceUnsupported, "", "", nil
	}
	if len(browsers) == 1 {
		return true, CookieSourceBrowser, browsers[0], "", nil
	}
	if len(files) == 1 {
		return true, CookieSourceFile, "", files[0], nil
	}
	return true, CookieSourceUnsupported, "", "", nil
}

func upsertManagedCookieLines(content, managedLine string) (string, error) {
	var kept []string
	sc := bufio.NewScanner(strings.NewReader(content))
	for sc.Scan() {
		line := sc.Text()
		trim := strings.TrimSpace(line)
		if trim == "" {
			kept = append(kept, line)
			continue
		}
		if strings.HasPrefix(trim, "#") || strings.HasPrefix(trim, ";") {
			kept = append(kept, line)
			continue
		}
		if opt, _, ok := splitConfigOption(trim); ok && (opt == "--cookies" || opt == "--cookies-from-browser") {
			continue
		}
		kept = append(kept, line)
	}
	if err := sc.Err(); err != nil {
		return "", err
	}

	// Drop trailing empty lines before append for stable files.
	for len(kept) > 0 && strings.TrimSpace(kept[len(kept)-1]) == "" {
		kept = kept[:len(kept)-1]
	}
	if managedLine != "" {
		kept = append(kept, managedLine)
	}
	if len(kept) == 0 {
		return "", nil
	}
	return strings.Join(kept, "\n") + "\n", nil
}

func splitConfigOption(line string) (opt, rest string, ok bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "--") {
		return "", "", false
	}
	// --opt=value only when '=' is not preceded by whitespace (paths may contain '=').
	if i := strings.IndexByte(line, '='); i > 0 && !strings.ContainsAny(line[:i], " \t") {
		return line[:i], unquoteConfigArg(line[i+1:]), true
	}
	end := len(line)
	for i, r := range line {
		if unicode.IsSpace(r) {
			end = i
			break
		}
	}
	opt = line[:end]
	rest = unquoteConfigArg(strings.TrimSpace(line[end:]))
	return opt, rest, true
}

func unquoteConfigArg(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func quoteConfigArg(s string) string {
	if strings.IndexFunc(s, unicode.IsSpace) >= 0 {
		return `"` + s + `"`
	}
	return s
}

// WrapCookieLinkageAPIError maps sentinel Cookie linkage errors to stable i18n keys; other errors pass through.
func WrapCookieLinkageAPIError(err error) error {
	if err == nil {
		return nil
	}
	if key := cookieLinkageErrorI18nKey(err); key != "" {
		return errors.New(key)
	}
	log.Printf("ytdlp cookie linkage: %v", err)
	return err
}

func cookieLinkageErrorI18nKey(err error) string {
	switch {
	case errors.Is(err, ErrCookieLinkageRiskAckRequired):
		return "errorRiskAckRequired"
	case errors.Is(err, ErrCookieLinkageUnsupportedPlatform):
		return "errorUnsupportedPlatform"
	case errors.Is(err, ErrCookieLinkageCookiesFileMissing):
		return "errorCookiesFileMissing"
	case errors.Is(err, ErrCookieLinkageInvalidBrowser):
		return "errorInvalidBrowser"
	default:
		return ""
	}
}
