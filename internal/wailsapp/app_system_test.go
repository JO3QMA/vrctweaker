package wailsapp

import (
	"runtime"
	"testing"
)

func TestGetLogicalProcessorCount(t *testing.T) {
	got, err := (&App{}).GetLogicalProcessorCount()
	if err != nil {
		t.Fatal(err)
	}
	want := runtime.NumCPU()
	if want < 1 {
		want = 1
	}
	if got != want {
		t.Fatalf("got %d want %d", got, want)
	}
}
