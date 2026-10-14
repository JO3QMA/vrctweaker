package usecase

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"vrchat-tweaker/internal/appversion"
	gh "vrchat-tweaker/internal/infrastructure/github"
)

type memSettings struct {
	m map[string]string
}

func (s *memSettings) Get(_ context.Context, key string) (string, error) {
	return s.m[key], nil
}

func (s *memSettings) Set(_ context.Context, key, value string) error {
	s.m[key] = value
	return nil
}

func (s *memSettings) GetAll(_ context.Context) (map[string]string, error) {
	out := make(map[string]string, len(s.m))
	for k, v := range s.m {
		out[k] = v
	}
	return out, nil
}

func TestAppUpdateUseCase_CheckForUpdate_promptsWhenNewer(t *testing.T) {
	const sha = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		payload := fmt.Sprintf(`{
			"tag_name":"v9.0.0",
			"prerelease":false,
			"draft":false,
			"html_url":"https://example.com",
			"body":"**SHA256 (zip):** `+"`%s`"+`",
			"assets":[{"name":"vrchat-tweaker-v9.0.0-windows-amd64.zip","browser_download_url":"https://github.com/JO3QMA/vrctweaker/releases/download/v9.0.0/vrchat-tweaker-v9.0.0-windows-amd64.zip"}]
		}`, sha)
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	appversion.SetVersion("1.0.0")
	settingsRepo := &memSettings{m: map[string]string{}}
	settings := NewSettingsUseCase(settingsRepo)
	uc := NewAppUpdateUseCase(settings, &gh.ReleaseClient{
		HTTPClient:        srv.Client(),
		ReleasesLatestURL: srv.URL,
		SkipURLValidation: false,
	}, nil)

	st, err := uc.CheckForUpdate(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if !st.UpdateAvailable || !st.ShouldPrompt || st.LatestVersion != "9.0.0" {
		t.Fatalf("status %+v", st)
	}
}

func TestAppUpdateUseCase_CheckForUpdate_respectsDismissedWithVPrefix(t *testing.T) {
	const sha = "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		payload := fmt.Sprintf(`{
			"tag_name":"v2.0.0",
			"prerelease":false,
			"draft":false,
			"html_url":"https://github.com/JO3QMA/vrctweaker/releases/tag/v2.0.0",
			"body":"**SHA256 (zip):** `+"`%s`"+`",
			"assets":[{"name":"vrchat-tweaker-v2.0.0-windows-amd64.zip","browser_download_url":"https://github.com/JO3QMA/vrctweaker/releases/download/v2.0.0/vrchat-tweaker-v2.0.0-windows-amd64.zip"}]
		}`, sha)
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	appversion.SetVersion("1.0.0")
	settingsRepo := &memSettings{m: map[string]string{keyDismissedUpdateVersion: "v2.0.0"}}
	settings := NewSettingsUseCase(settingsRepo)
	uc := NewAppUpdateUseCase(settings, &gh.ReleaseClient{
		HTTPClient:        srv.Client(),
		ReleasesLatestURL: srv.URL,
		SkipURLValidation: false,
	}, nil)

	st, err := uc.CheckForUpdate(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if !st.UpdateAvailable || st.ShouldPrompt {
		t.Fatalf("status %+v", st)
	}
}

func TestAppUpdateUseCase_CheckForUpdate_respectsDismissed(t *testing.T) {
	const sha = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		payload := fmt.Sprintf(`{
			"tag_name":"v2.0.0",
			"prerelease":false,
			"draft":false,
			"html_url":"https://example.com",
			"body":"**SHA256 (zip):** `+"`%s`"+`",
			"assets":[{"name":"vrchat-tweaker-v2.0.0-windows-amd64.zip","browser_download_url":"https://github.com/JO3QMA/vrctweaker/releases/download/v2.0.0/vrchat-tweaker-v2.0.0-windows-amd64.zip"}]
		}`, sha)
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	appversion.SetVersion("1.0.0")
	settingsRepo := &memSettings{m: map[string]string{keyDismissedUpdateVersion: "2.0.0"}}
	settings := NewSettingsUseCase(settingsRepo)
	uc := NewAppUpdateUseCase(settings, &gh.ReleaseClient{
		HTTPClient:        srv.Client(),
		ReleasesLatestURL: srv.URL,
		SkipURLValidation: false,
	}, nil)

	st, err := uc.CheckForUpdate(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if !st.UpdateAvailable || st.ShouldPrompt {
		t.Fatalf("status %+v", st)
	}
}

func TestAppUpdateUseCase_ApplyUpdate_nonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows-only negative test")
	}
	uc := NewAppUpdateUseCase(nil, nil, nil)
	err := uc.ApplyUpdate(context.Background(), "1.0.0")
	if err == nil || err.Error() != ErrAppUpdateUnsupportedSent {
		t.Fatalf("err %v", err)
	}
}
