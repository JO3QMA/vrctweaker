//go:build windows

package powerplan

import (
	"os/exec"
	"syscall"
)

// CREATE_NO_WINDOW — avoid flashing a console when a GUI app runs powercfg.
const createNoWindow = 0x08000000

func hideConsoleWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNoWindow,
	}
}
