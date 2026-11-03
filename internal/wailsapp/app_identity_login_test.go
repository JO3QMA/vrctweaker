package wailsapp

import (
	"testing"
)

func TestApp_Login_beforeReady(t *testing.T) {
	a := &App{ctx: t.Context()}
	res := a.Login("u", "p", "")
	if res.OK {
		t.Fatal("expected failure")
	}
	if res.Error != errAppNotInitialized.Error() {
		t.Fatalf("error = %q", res.Error)
	}
}
