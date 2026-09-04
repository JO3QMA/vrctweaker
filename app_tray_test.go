package main

import "testing"

func TestApp_closeToTrayEffective_defaultsWithoutTray(t *testing.T) {
	a := &App{}
	if a.closeToTrayEffective(t.Context()) {
		t.Fatal("want false without settings/tray")
	}
}

func TestApp_handleBeforeClose_withoutTray(t *testing.T) {
	a := &App{}
	if a.handleBeforeClose(t.Context()) {
		t.Fatal("want false without close-to-tray")
	}
}
