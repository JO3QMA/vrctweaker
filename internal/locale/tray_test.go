package locale

import "testing"

func TestTrayMenuLabels_ja(t *testing.T) {
	show, quit := TrayMenuLabels("ja")
	if show == "" || quit == "" {
		t.Fatal("want non-empty labels")
	}
}

func TestTrayMenuLabels_defaultEnglish(t *testing.T) {
	show, quit := TrayMenuLabels("en")
	if show != "Show window" || quit != "Quit" {
		t.Fatalf("got %q / %q", show, quit)
	}
}
