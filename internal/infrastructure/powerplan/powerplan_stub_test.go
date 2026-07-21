//go:build !windows

package powerplan

import (
	"context"
	"testing"
)

func TestListDetected_stub(t *testing.T) {
	plans, err := ListDetected(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(plans) != 0 {
		t.Fatalf("want empty, got %d", len(plans))
	}
}

func TestSetActive_stub(t *testing.T) {
	if err := SetActive(context.Background(), "x"); err == nil {
		t.Fatal("want error")
	}
}

func TestResolvePreset_stub(t *testing.T) {
	if _, err := ResolvePreset(context.Background(), "balanced"); err == nil {
		t.Fatal("want error")
	}
}
