package github

import "testing"

func TestValidateReleaseDownloadURL_githubPath(t *testing.T) {
	ok := "https://github.com/JO3QMA/vrctweaker/releases/download/v1.0.0/vrchat-tweaker-v1.0.0-windows-amd64.zip"
	if err := ValidateReleaseDownloadURL(ok); err != nil {
		t.Fatal(err)
	}
	bad := "https://github.com/other/repo/releases/download/v1.0.0/evil.zip"
	if err := ValidateReleaseDownloadURL(bad); err == nil {
		t.Fatal("expected path rejection")
	}
}

func TestValidateReleasePageURL(t *testing.T) {
	if err := ValidateReleasePageURL("https://github.com/JO3QMA/vrctweaker/releases/latest"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateReleasePageURL("file:///etc/passwd"); err == nil {
		t.Fatal("expected error")
	}
	if err := ValidateReleasePageURL("https://github.com/JO3QMA/vrctweaker/releasesX"); err == nil {
		t.Fatal("expected segment boundary rejection")
	}
	if err := ValidateReleasePageURL("https://github.com:8443/JO3QMA/vrctweaker/releases/latest"); err == nil {
		t.Fatal("expected port rejection")
	}
}

func TestValidateHostedDownloadURL_repoPrefix(t *testing.T) {
	const ytdlpURL = "https://github.com/yt-dlp/yt-dlp/releases/download/2026.07.04/yt-dlp.exe"
	ytdlpOpt := HostedDownloadOptions{RepoPathPrefix: "/yt-dlp/yt-dlp/releases/download/"}
	if err := ValidateHostedDownloadURL(ytdlpURL, ytdlpOpt); err != nil {
		t.Fatal(err)
	}
	vrct := "https://github.com/JO3QMA/vrctweaker/releases/download/v1.0.0/a.zip"
	if err := ValidateHostedDownloadURL(vrct, ytdlpOpt); err == nil {
		t.Fatal("expected other repo to be rejected")
	}
	if err := ValidateHostedDownloadURL(vrct, HostedDownloadOptions{}); err != nil {
		t.Fatal(err)
	}
	withPort := "https://github.com:443/yt-dlp/yt-dlp/releases/download/2026.07.04/yt-dlp.exe"
	if err := ValidateHostedDownloadURL(withPort, HostedDownloadOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateHostedDownloadURL(withPort, HostedDownloadOptions{RejectPort: true}); err == nil {
		t.Fatal("expected port rejection")
	}
	cdn := "https://objects.githubusercontent.com/x/yt-dlp.exe"
	if err := ValidateHostedDownloadURL(cdn, ytdlpOpt); err != nil {
		t.Fatal(err)
	}
}

func TestValidateReleaseDownloadURL_segmentBoundary(t *testing.T) {
	bad := "https://github.com/JO3QMA/vrctweaker/releases/downloadX/v1.0.0/evil.zip"
	if err := ValidateReleaseDownloadURL(bad); err == nil {
		t.Fatal("expected path rejection")
	}
}
