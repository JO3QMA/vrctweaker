//go:build windows

package vrchatwindow

import (
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modUser32             = windows.NewLazySystemDLL("user32.dll")
	procGetWindowRect     = modUser32.NewProc("GetWindowRect")
	procSetWindowPos      = modUser32.NewProc("SetWindowPos")
	procShowWindow        = modUser32.NewProc("ShowWindow")
	procIsZoomed          = modUser32.NewProc("IsZoomed")
	procMonitorFromWindow = modUser32.NewProc("MonitorFromWindow")
	procGetMonitorInfoW   = modUser32.NewProc("GetMonitorInfoW")
	procGetWindow         = modUser32.NewProc("GetWindow")
)

const (
	swRestore             = 9
	swpNoZOrder           = 0x0004
	swpNoActivate         = 0x0010
	monitorDefaultNearest = 2
	gwOwner               = 4
)

type rect struct {
	Left, Top, Right, Bottom int32
}

type monitorInfo struct {
	CbSize    uint32
	RcMonitor rect
	RcWork    rect
	DwFlags   uint32
}

func resize(width, height int) error {
	pids, err := vrchatPIDs()
	if err != nil {
		return err
	}
	if len(pids) == 0 {
		return ErrNotRunning
	}
	hwnd, ok := findMainWindow(pids)
	if !ok {
		return ErrNoWindow
	}
	if isZoomed(hwnd) {
		showWindow(hwnd, swRestore)
	}
	var before rect
	if err := getWindowRect(hwnd, &before); err != nil {
		return err
	}
	if err := setWindowPos(hwnd, before.Left, before.Top, int32(width), int32(height)); err != nil {
		// Exclusive fullscreen often rejects SetWindowPos — treat as skip.
		if coversMonitor(hwnd) {
			return nil
		}
		return err
	}
	var after rect
	if err := getWindowRect(hwnd, &after); err != nil {
		return err
	}
	gotW := int(after.Right - after.Left)
	gotH := int(after.Bottom - after.Top)
	if gotW == width && gotH == height {
		return nil
	}
	// Size unchanged while still covering the monitor → exclusive fullscreen skip.
	if coversMonitor(hwnd) {
		return nil
	}
	return ErrResizeFailed
}

func vrchatPIDs() ([]uint32, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)

	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	if err = windows.Process32First(snapshot, &pe); err != nil {
		return nil, err
	}
	var pids []uint32
	for {
		name := windows.UTF16ToString(pe.ExeFile[:])
		if strings.EqualFold(name, "VRChat.exe") {
			pids = append(pids, pe.ProcessID)
		}
		err = windows.Process32Next(snapshot, &pe)
		if err != nil {
			if err == windows.ERROR_NO_MORE_FILES {
				break
			}
			return nil, err
		}
	}
	return pids, nil
}

func findMainWindow(pids []uint32) (windows.HWND, bool) {
	want := make(map[uint32]struct{}, len(pids))
	for _, p := range pids {
		want[p] = struct{}{}
	}
	type cand struct {
		hwnd windows.HWND
		area int64
	}
	var best cand
	cb := syscall.NewCallback(func(hwnd windows.HWND, _ uintptr) uintptr {
		if !windows.IsWindowVisible(hwnd) {
			return 1
		}
		// Skip owned windows (dialogs); prefer top-level.
		owner, _, _ := procGetWindow.Call(uintptr(hwnd), gwOwner)
		if owner != 0 {
			return 1
		}
		var pid uint32
		_, _ = windows.GetWindowThreadProcessId(hwnd, &pid)
		if _, ok := want[pid]; !ok {
			return 1
		}
		var r rect
		if getWindowRect(hwnd, &r) != nil {
			return 1
		}
		w := int64(r.Right - r.Left)
		h := int64(r.Bottom - r.Top)
		if w <= 0 || h <= 0 {
			return 1
		}
		area := w * h
		if area > best.area {
			best = cand{hwnd: hwnd, area: area}
		}
		return 1
	})
	_ = windows.EnumWindows(cb, unsafe.Pointer(nil))
	return best.hwnd, best.hwnd != 0
}

func coversMonitor(hwnd windows.HWND) bool {
	var wr rect
	if getWindowRect(hwnd, &wr) != nil {
		return false
	}
	mon, _, _ := procMonitorFromWindow.Call(uintptr(hwnd), monitorDefaultNearest)
	if mon == 0 {
		return false
	}
	var mi monitorInfo
	mi.CbSize = uint32(unsafe.Sizeof(mi))
	r1, _, _ := procGetMonitorInfoW.Call(mon, uintptr(unsafe.Pointer(&mi)))
	if r1 == 0 {
		return false
	}
	return wr.Left == mi.RcMonitor.Left &&
		wr.Top == mi.RcMonitor.Top &&
		wr.Right == mi.RcMonitor.Right &&
		wr.Bottom == mi.RcMonitor.Bottom
}

func isZoomed(hwnd windows.HWND) bool {
	r, _, _ := procIsZoomed.Call(uintptr(hwnd))
	return r != 0
}

func showWindow(hwnd windows.HWND, cmd int32) {
	procShowWindow.Call(uintptr(hwnd), uintptr(cmd))
}

func getWindowRect(hwnd windows.HWND, r *rect) error {
	ret, _, err := procGetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(r)))
	if ret == 0 {
		return err
	}
	return nil
}

func setWindowPos(hwnd windows.HWND, x, y, cx, cy int32) error {
	const flags = swpNoZOrder | swpNoActivate
	ret, _, err := procSetWindowPos.Call(
		uintptr(hwnd),
		0,
		uintptr(x),
		uintptr(y),
		uintptr(cx),
		uintptr(cy),
		flags,
	)
	if ret == 0 {
		return err
	}
	return nil
}
