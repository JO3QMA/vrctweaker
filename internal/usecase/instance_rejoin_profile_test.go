package usecase

import (
	"testing"

	"vrchat-tweaker/internal/domain/launcher"
)

func TestResolveInstanceRejoinProfile_lastLaunch(t *testing.T) {
	profiles := []*launcher.LaunchProfile{
		{ID: "a", Name: "A"},
		{ID: "b", Name: "B", IsDefault: true},
	}
	got := ResolveInstanceRejoinProfileID(profiles, "a")
	if got != "a" {
		t.Fatalf("got %q want a", got)
	}
}

func TestResolveInstanceRejoinProfile_staleLastLaunch(t *testing.T) {
	profiles := []*launcher.LaunchProfile{
		{ID: "a", Name: "A"},
		{ID: "b", Name: "B", IsDefault: true},
	}
	got := ResolveInstanceRejoinProfileID(profiles, "deleted")
	if got != "b" {
		t.Fatalf("got %q want default b", got)
	}
}

func TestResolveInstanceRejoinProfile_firstProfile(t *testing.T) {
	profiles := []*launcher.LaunchProfile{
		{ID: "a", Name: "A"},
		{ID: "b", Name: "B"},
	}
	got := ResolveInstanceRejoinProfileID(profiles, "")
	if got != "a" {
		t.Fatalf("got %q want first a", got)
	}
}

func TestResolveInstanceRejoinProfile_noProfiles(t *testing.T) {
	if got := ResolveInstanceRejoinProfileID(nil, "x"); got != "" {
		t.Fatalf("got %q want empty", got)
	}
}
