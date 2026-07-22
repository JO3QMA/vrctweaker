//go:build windows

package vrchatwindow

import (
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modUser32          = windows.NewLazySystemDLL("user32.dll")
	procGetWindowRect  = modUser32.NewProc("GetWindowRect")
	procSetWindowPos   = modUser32.NewProc("SetWindowPos")
	procShowWindow     = modUser32.NewProc("ShowWindow")
	procIsZoomed       = modUser32.NewProc("IsZoomed")
	procGetWindow      = modUser32.NewProc("GetWindow")
	procGetClassNameW  = modUser32.NewProc("GetClassNameW")
	procGetWindowTextW = modUser32.NewProc("GetWindowTextW")

	resizeMu sync.Mutex

	// Registered once — NewCallback allocates permanently.
	enumWindowsCB = sync.OnceValue(func() uintptr {
		return syscall.NewCallback(enumWindowsProc)
	})
)

const (
	swRestore     = 9
	swpNoZOrder   = 0x0004
	swpNoActivate = 0x0010
	gwOwner       = 4

	minWindowArea       = int64(640 * 480)
	restorePollInterval = 10 * time.Millisecond
	restorePollTimeout  = 500 * time.Millisecond
)

type rect struct {
	Left, Top, Right, Bottom int32
}

type windowCand struct {
	hwnd windows.HWND
	area int64
}

type enumData struct {
	want      map[uint32]struct{}
	bestByPID map[uint32]windowCand
}

func resize(width, height int) error {
	if err := unmaximizeIfNeeded(); err != nil {
		return err
	}

	resizeMu.Lock()
	defer resizeMu.Unlock()

	hwnd, err := resolveMainHWND()
	if err != nil {
		return err
	}
	if isZoomed(hwnd) {
		return ErrResizeFailed
	}
	return applySize(hwnd, width, height)
}

// unmaximizeIfNeeded restores a maximized main window. The poll waits without holding resizeMu.
func unmaximizeIfNeeded() error {
	resizeMu.Lock()
	hwnd, err := resolveMainHWND()
	if err != nil {
		resizeMu.Unlock()
		return err
	}
	if !isZoomed(hwnd) {
		resizeMu.Unlock()
		return nil
	}
	// ShowWindow return is previous visibility, not success/failure.
	showWindow(hwnd, swRestore)
	resizeMu.Unlock()
	return waitUntilNotZoomed(hwnd)
}

func applySize(hwnd windows.HWND, width, height int) error {
	var before rect
	if err := getWindowRect(hwnd, &before); err != nil {
		return err
	}
	if err := setWindowPos(hwnd, before.Left, before.Top, int32(width), int32(height)); err != nil {
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
	return ErrResizeFailed
}

func waitUntilNotZoomed(hwnd windows.HWND) error {
	deadline := time.Now().Add(restorePollTimeout)
	for {
		if !isZoomed(hwnd) {
			return nil
		}
		if time.Now().After(deadline) {
			return ErrResizeFailed
		}
		time.Sleep(restorePollInterval)
	}
}

func resolveMainHWND() (windows.HWND, error) {
	pids, err := vrchatPIDs()
	if err != nil {
		return 0, err
	}
	if len(pids) == 0 {
		return 0, ErrNotRunning
	}
	return findMainWindow(pids)
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

func findMainWindow(pids []uint32) (windows.HWND, error) {
	want := make(map[uint32]struct{}, len(pids))
	for _, p := range pids {
		want[p] = struct{}{}
	}
	data := &enumData{
		want:      want,
		bestByPID: make(map[uint32]windowCand),
	}
	_ = windows.EnumWindows(enumWindowsCB(), unsafe.Pointer(data))
	switch len(data.bestByPID) {
	case 0:
		return 0, ErrNoWindow
	case 1:
		for _, c := range data.bestByPID {
			return c.hwnd, nil
		}
		return 0, ErrNoWindow
	default:
		return 0, ErrMultipleInstances
	}
}

func enumWindowsProc(hwnd windows.HWND, lparam uintptr) uintptr {
	if lparam == 0 {
		return 1
	}
	data := (*enumData)(unsafe.Pointer(lparam))
	if !windows.IsWindowVisible(hwnd) {
		return 1
	}
	// GetWindow: ignore Call's last-error — it is often stale on success.
	owner, _, _ := procGetWindow.Call(uintptr(hwnd), gwOwner)
	if owner != 0 {
		return 1
	}
	if !isLikelyVRChatGameWindow(hwnd) {
		return 1
	}
	var pid uint32
	_, _ = windows.GetWindowThreadProcessId(hwnd, &pid)
	if _, ok := data.want[pid]; !ok {
		return 1
	}
	var r rect
	if getWindowRect(hwnd, &r) != nil {
		return 1
	}
	w := int64(r.Right - r.Left)
	h := int64(r.Bottom - r.Top)
	area := w * h
	if area < minWindowArea {
		return 1
	}
	if prev, ok := data.bestByPID[pid]; !ok || area > prev.area {
		data.bestByPID[pid] = windowCand{hwnd: hwnd, area: area}
	}
	return 1
}

func isLikelyVRChatGameWindow(hwnd windows.HWND) bool {
	return classOrTitleLooksLikeVRChat(windowClassName(hwnd), windowTitle(hwnd))
}

func windowClassName(hwnd windows.HWND) string {
	buf := make([]uint16, 256)
	n, _, _ := procGetClassNameW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return ""
	}
	return windows.UTF16ToString(buf[:n])
}

func windowTitle(hwnd windows.HWND) string {
	buf := make([]uint16, 256)
	n, _, _ := procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return ""
	}
	return windows.UTF16ToString(buf[:n])
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
		uintptr(flags),
	)
	if ret == 0 {
		return err
	}
	return nil
}
