package wailsapp

import (
	"encoding/json"
	"testing"
)

func TestGetAutomationRuntimeStatus_notInitializedJSON(t *testing.T) {
	got, err := (&App{}).GetAutomationRuntimeStatus()
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	b, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	const want = `{"available":false,"reasonKey":"not_initialized"}`
	if string(b) != want {
		t.Fatalf("json=%s", b)
	}
}
