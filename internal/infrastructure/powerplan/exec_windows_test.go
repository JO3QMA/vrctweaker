//go:build windows

package powerplan

import (
	"os/exec"
	"testing"
)

func TestHideConsoleWindow_setsNoWindowFlags(t *testing.T) {
	cmd := exec.Command("powercfg", "/list")
	hideConsoleWindow(cmd)
	if cmd.SysProcAttr == nil {
		t.Fatal("SysProcAttr nil")
	}
	if !cmd.SysProcAttr.HideWindow {
		t.Fatal("HideWindow want true")
	}
	if cmd.SysProcAttr.CreationFlags&createNoWindow == 0 {
		t.Fatalf("CreationFlags=%#x missing CREATE_NO_WINDOW", cmd.SysProcAttr.CreationFlags)
	}
}
