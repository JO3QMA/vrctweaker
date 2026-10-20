package appupdate

import (
	"net/http"
	"testing"
)

func TestValidateReleaseDownloadRedirect_rejectsDisallowedHost(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://evil.example/payload.zip", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateReleaseDownloadRedirect(req, []*http.Request{{}}); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateReleaseDownloadRedirect_allowsGitHubAsset(t *testing.T) {
	url := "https://github.com/JO3QMA/vrctweaker/releases/download/v1.0.0/vrchat-tweaker-v1.0.0-windows-amd64.zip"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateReleaseDownloadRedirect(req, []*http.Request{{}}); err != nil {
		t.Fatal(err)
	}
}
