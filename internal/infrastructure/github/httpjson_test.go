package github

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetJSON_ytdlpHeadersAndStatusText(t *testing.T) {
	t.Parallel()
	var gotUA, gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotAccept = r.Header.Get("Accept")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"API rate limit exceeded"}`))
	}))
	t.Cleanup(srv.Close)

	_, err := GetJSON(context.Background(), srv.Client(), srv.URL, "VRChat Tweaker/test", APIJSONYTDLP)
	if gotUA != "VRChat Tweaker/test" {
		t.Fatalf("User-Agent %q", gotUA)
	}
	if gotAccept != githubAPIAccept {
		t.Fatalf("Accept %q", gotAccept)
	}
	const want = `github api: 403 Forbidden: {"message":"API rate limit exceeded"}`
	if err == nil || err.Error() != want {
		t.Fatalf("err %v", err)
	}
}

func TestGetJSON_ytdlpTruncatesStatusBody(t *testing.T) {
	t.Parallel()
	body := strings.Repeat("x", 250)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	_, err := GetJSON(context.Background(), srv.Client(), srv.URL, "ua", APIJSONYTDLP)
	want := "github api: 500 Internal Server Error: " + strings.Repeat("x", 200) + "…"
	if err == nil || err.Error() != want {
		t.Fatalf("err %v", err)
	}
}

func TestReadYTDLPReleaseBody_truncatesWithoutError(t *testing.T) {
	t.Parallel()
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader("abcdefghij")),
	}
	body, err := readYTDLPReleaseBody(resp, 4)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "abcd" {
		t.Fatalf("body %q", body)
	}
}

func TestGetJSON_appReleaseStatusAndOversize(t *testing.T) {
	t.Parallel()
	statusSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != githubAPIAccept {
			t.Errorf("Accept %q", r.Header.Get("Accept"))
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("  missing  \n"))
	}))
	t.Cleanup(statusSrv.Close)

	_, err := GetJSON(context.Background(), statusSrv.Client(), statusSrv.URL, "ua", APIJSONAppRelease)
	if err == nil || err.Error() != "github releases: HTTP 404: missing" {
		t.Fatalf("status err %v", err)
	}

	overSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, strings.Repeat("a", int(appReleaseMaxBytes)+1))
	}))
	t.Cleanup(overSrv.Close)

	_, err = GetJSON(context.Background(), overSrv.Client(), overSrv.URL, "ua", APIJSONAppRelease)
	if err == nil || err.Error() != "github releases: response exceeds size limit" {
		t.Fatalf("oversize err %v", err)
	}
}

func TestGetJSON_requestErrorPrefixes(t *testing.T) {
	t.Parallel()
	client := &http.Client{Transport: roundTripErr{}}
	_, yErr := GetJSON(context.Background(), client, "https://example.com/latest", "ua", APIJSONYTDLP)
	if yErr == nil || !strings.HasPrefix(yErr.Error(), "github api request:") || !strings.Contains(yErr.Error(), "boom") {
		t.Fatalf("ytdlp %v", yErr)
	}
	_, aErr := GetJSON(context.Background(), client, "https://example.com/latest", "ua", APIJSONAppRelease)
	if aErr == nil || !strings.HasPrefix(aErr.Error(), "github releases:") || !strings.Contains(aErr.Error(), "boom") {
		t.Fatalf("app %v", aErr)
	}
}

func TestGetJSON_unknownKind(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	_, err := GetJSON(context.Background(), srv.Client(), srv.URL, "ua", 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTruncateErrBody(t *testing.T) {
	t.Parallel()
	if got := TruncateErrBody("  hi  ", 10); got != "hi" {
		t.Fatalf("got %q", got)
	}
	if got := TruncateErrBody(strings.Repeat("あ", 3), 2); got != "ああ…" {
		t.Fatalf("got %q", got)
	}
}

type roundTripErr struct{}

func (roundTripErr) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("boom")
}
