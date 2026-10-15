//go:build windows

package appupdate

import (
	"os/exec"
	"testing"

	"golang.org/x/sys/windows"
)

func TestSetUpdateHelperProcessAttrs_breakawayFromJob(t *testing.T) {
	cmd := exec.Command("powershell.exe", "-NoProfile")
	setUpdateHelperProcessAttrs(cmd)
	if cmd.SysProcAttr == nil {
		t.Fatal("SysProcAttr nil")
	}
	want := windows.CREATE_NEW_PROCESS_GROUP | windows.CREATE_BREAKAWAY_FROM_JOB
	if cmd.SysProcAttr.CreationFlags != want {
		t.Fatalf("CreationFlags=%#x want %#x", cmd.SysProcAttr.CreationFlags, want)
	}
}
