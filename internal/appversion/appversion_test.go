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

func TestValidateProductVersion(t *testing.T) {
	if err := ValidateProductVersion("1.0.0"); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"", "../x", "a/b", "1..2"} {
		if err := ValidateProductVersion(bad); err == nil {
			t.Fatalf("expected error for %q", bad)
		}
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
