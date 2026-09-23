package appversion

import "testing"

func TestUserAgent(t *testing.T) {
	SetVersion("1.2.3")
	got := UserAgent()
	want := "VRChat Tweaker/1.2.3 (+https://github.com/JO3QMA/vrctweaker/issues)"
	if got != want {
		t.Fatalf("UserAgent() = %q, want %q", got, want)
	}
}

func TestParseProductVersionFromWailsJSON(t *testing.T) {
	const sample = `{
  "name": "vrchat-tweaker",
  "info": {
    "productName": "VRChat Tweaker",
    "productVersion": "0.1.0"
  }
}`
	got, err := ParseProductVersionFromWailsJSON([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	if got != "0.1.0" {
		t.Fatalf("version %q", got)
	}
}

func TestParseProductVersionFromWailsJSON_missing(t *testing.T) {
	got, err := ParseProductVersionFromWailsJSON([]byte(`{"info":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}
