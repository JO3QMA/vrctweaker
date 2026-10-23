/*
 * Copyright (C) 2019 The Systray Authors. All Rights Reserved.
 */

// Package w32systray is a Windows notification-area helper forked from github.com/tadvi/systray (MIT).
package w32systray

import (
	"errors"
	"log"
	"path/filepath"
	"syscall"
	"unsafe"
)

const (
	WM_LBUTTONUP     = 0x0202
	WM_LBUTTONDBLCLK = 0x0203
	WM_RBUTTONUP     = 0x0205
	WM_USER          = 0x0400
	WM_TRAYICON      = WM_USER + 69

	WS_EX_APPWINDOW     = 0x00040000
	WS_OVERLAPPEDWINDOW = 0x00000000 | 0x00C00000 | 0x00080000 | 0x00040000 | 0x00020000 | 0x00010000
	CW_USEDEFAULT       = 0x80000000

	NIM_ADD        = 0x00000000
	NIM_MODIFY     = 0x00000001
	NIM_DELETE     = 0x00000002
	NIM_SETVERSION = 0x00000004

	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004
	NIF_STATE   = 0x00000008
	NIF_INFO    = 0x00000010

	NIS_HIDDEN = 0x00000001

	NIIF_NONE               = 0x00000000
	NIIF_INFO               = 0x00000001
	NIIF_WARNING            = 0x00000002
	NIIF_ERROR              = 0x00000003
	NIIF_USER               = 0x00000004
	NIIF_NOSOUND            = 0x00000010
	NIIF_LARGE_ICON         = 0x00000020
	NIIF_RESPECT_QUIET_TIME = 0x00000080
	NIIF_ICON_MASK          = 0x0000000F

	IMAGE_BITMAP    = 0
	IMAGE_ICON      = 1
	LR_LOADFROMFILE = 0x00000010
	LR_DEFAULTSIZE  = 0x00000040

	IDC_ARROW     = 32512
	COLOR_WINDOW  = 5
	COLOR_BTNFACE = 15

	GWLP_USERDATA       = -21
	WS_CLIPSIBLINGS     = 0x04000000
	WS_EX_CONTROLPARENT = 0x00010000

	HWND_MESSAGE       = ^HWND(2)
	NOTIFYICON_VERSION = 4

	IDI_APPLICATION = 32512
	WM_APP          = 32768
	WM_COMMAND      = 273

	MenuItemMsgID       = WM_APP + 1024
	NotifyIconMessageId = WM_APP + 1

	// WMTrayShutdown is posted from other goroutines to tear down the icon on the tray thread.
	WMTrayShutdown = WM_APP + 2001

	MF_STRING       = 0x00000000
	MF_ENABLED      = 0x00000000
	MF_GRAYED       = 0x00000001
	MF_DISABLED     = 0x00000002
	MF_SEPARATOR    = 0x00000800
	MF_CHECKED      = 0x00000008
	MF_MENUBARBREAK = 0x00000020

	TPM_LEFTALIGN = 0x0000
	WM_NULL       = 0
	PM_NOREMOVE   = 0x0000
)

var (
	kernel32         = syscall.MustLoadDLL("kernel32")
	GetModuleHandle  = kernel32.MustFindProc("GetModuleHandleW")
	GetConsoleWindow = kernel32.MustFindProc("GetConsoleWindow")
	GetLastError     = kernel32.MustFindProc("GetLastError")

	shell32          = syscall.MustLoadDLL("shell32.dll")
	Shell_NotifyIcon = shell32.MustFindProc("Shell_NotifyIconW")

	user32 = syscall.MustLoadDLL("user32.dll")

	GetMessage       = user32.MustFindProc("GetMessageW")
	PeekMessage      = user32.MustFindProc("PeekMessageW")
	IsDialogMessage  = user32.MustFindProc("IsDialogMessageW")
	TranslateMessage = user32.MustFindProc("TranslateMessage")
	DispatchMessage  = user32.MustFindProc("DispatchMessageW")

	ShowWindow       = user32.MustFindProc("ShowWindow")
	UpdateWindow     = user32.MustFindProc("UpdateWindow")
	DefWindowProc    = user32.MustFindProc("DefWindowProcW")
	RegisterClassEx  = user32.MustFindProc("RegisterClassExW")
	GetDesktopWindow = user32.MustFindProc("GetDesktopWindow")
	CreateWindowEx   = user32.MustFindProc("CreateWindowExW")

	CreatePopupMenu           = user32.MustFindProc("CreatePopupMenu")
	procAppendMenuW           = user32.MustFindProc("AppendMenuW")
	procGetCursorPos          = user32.MustFindProc("GetCursorPos")
	procSetForegroundWindow   = user32.MustFindProc("SetForegroundWindow")
	procTrackPopupMenu        = user32.MustFindProc("TrackPopupMenu")
	procPostMessage           = user32.MustFindProc("PostMessageW")
	procPostThreadMessage     = user32.MustFindProc("PostThreadMessageW")
	procPostQuitMessage       = user32.MustFindProc("PostQuitMessage")
	procRegisterWindowMessage = user32.MustFindProc("RegisterWindowMessageW")
	procDestroyWindow         = user32.MustFindProc("DestroyWindow")
	procDestroyIcon           = user32.MustFindProc("DestroyIcon")
	procDestroyMenu           = user32.MustFindProc("DestroyMenu")
	procUnregisterClassW      = user32.MustFindProc("UnregisterClassW")

	LoadImage  = user32.MustFindProc("LoadImageW")
	LoadIcon   = user32.MustFindProc("LoadIconW")
	LoadCursor = user32.MustFindProc("LoadCursorW")
)

type NOTIFYICONDATA struct {
	CbSize           uint32
	HWnd             HWND
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            HICON
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         GUID
	HBalloonIcon     HICON
}

type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     HINSTANCE
	HIcon         HICON
	HCursor       HCURSOR
	HbrBackground HBRUSH
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       HICON
}

type MSG struct {
	HWnd    HWND
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type POINT struct {
	X, Y int32
}

type (
	HANDLE    uintptr
	HINSTANCE HANDLE
	HCURSOR   HANDLE
	HICON     HANDLE
	HWND      HANDLE
	HGDIOBJ   HANDLE
	HBRUSH    HGDIOBJ
)

type HMENU HANDLE

type WindowProc func(hwnd HWND, msg uint32, wparam, lparam uintptr) uintptr

type MenuItem struct {
	Label string

	Disabled  bool
	Checked   bool
	BarBreak  bool
	Separator bool

	OnClick func()
}

type Systray struct {
	id     uint32
	hwnd   HWND
	hinst  HINSTANCE
	lclick func()
	rclick func()

	Menu []*MenuItem

	wmTaskbarCreated uint32
	notifyClassName  *uint16
	iconHandle       HICON
	iconOwned        bool
	tooltip          string
	visible          bool
	destroyed        bool
}

func New() (*Systray, error) {
	ni := &Systray{lclick: func() {}, rclick: func() {}}

	const notifyIconClassName = "VRChatTweaker.NotifyIcon"
	ni.notifyClassName = syscall.StringToUTF16Ptr(notifyIconClassName)
	hinst, err := RegisterWindow(notifyIconClassName, ni.WinProc)
	if err != nil {
		return nil, err
	}
	ni.hinst = hinst

	taskbarCreated, _, _ := procRegisterWindowMessage.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("TaskbarCreated"))),
	)
	if taskbarCreated == 0 {
		return nil, errors.New("register TaskbarCreated message failed")
	}
	ni.wmTaskbarCreated = uint32(taskbarCreated)

	hwnd, _, _ := CreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(notifyIconClassName))),
		0,
		0,
		0,
		0,
		0,
		0,
		uintptr(HWND_MESSAGE),
		0,
		0,
		0)
	if hwnd == 0 {
		ni.destroyWindowClass()
		return nil, errors.New("create notify win failed")
	}

	ni.hwnd = HWND(hwnd)

	nid := NOTIFYICONDATA{
		HWnd:             HWND(hwnd),
		UFlags:           NIF_MESSAGE | NIF_STATE,
		DwState:          NIS_HIDDEN,
		DwStateMask:      NIS_HIDDEN,
		UCallbackMessage: NotifyIconMessageId,
	}
	nid.CbSize = uint32(unsafe.Sizeof(nid))

	ret, _, _ := Shell_NotifyIcon.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
	if ret == 0 {
		_ = ni.Destroy()
		return nil, errors.New("shell notify create failed")
	}

	nid.UVersion = NOTIFYICON_VERSION

	ret, _, _ = Shell_NotifyIcon.Call(NIM_SETVERSION, uintptr(unsafe.Pointer(&nid)))
	if ret == 0 {
		_, _, _ = Shell_NotifyIcon.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
		_ = ni.Destroy()
		return nil, errors.New("shell notify version failed")
	}

	ni.id = nid.UID
	return ni, nil
}

func (p *Systray) HWND() HWND {
	return p.hwnd
}

// AppendMenu add menu item.
func (p *Systray) AppendMenu(label string, onclick func()) {
	p.Menu = append(p.Menu, &MenuItem{Label: label, OnClick: onclick})
}

// AppendSeparator to the menu.
func (p *Systray) AppendSeparator() {
	p.Menu = append(p.Menu, &MenuItem{Separator: true})
}

// Stop removes the notification icon. Prefer Destroy for full resource cleanup.
func (p *Systray) Stop() error {
	return p.deleteNotifyIcon()
}

func (p *Systray) deleteNotifyIcon() error {
	if p.hwnd == 0 {
		return nil
	}
	nid := NOTIFYICONDATA{
		UID:  p.id,
		HWnd: HWND(p.hwnd),
	}
	nid.CbSize = uint32(unsafe.Sizeof(nid))

	ret, _, _ := Shell_NotifyIcon.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
	if ret == 0 {
		return errors.New("shell notify delete failed")
	}
	return nil
}

// EnsureMessageQueue creates the per-thread message queue before cross-thread posts.
func (p *Systray) EnsureMessageQueue() {
	var msg MSG
	_, _, _ = PeekMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0, PM_NOREMOVE)
}

// Destroy removes the tray icon and releases Win32 handles (HWND, owned HICON, window class).
func (p *Systray) Destroy() error {
	if p.destroyed {
		return nil
	}
	if err := p.deleteNotifyIcon(); err != nil {
		log.Printf("w32systray: notify icon delete: %v", err)
	}
	p.destroyed = true
	p.releaseOwnedIcon()
	if p.hwnd != 0 {
		_, _, _ = procDestroyWindow.Call(uintptr(p.hwnd))
		p.hwnd = 0
	}
	p.destroyWindowClass()
	return nil
}

func (p *Systray) releaseOwnedIcon() {
	if p.iconOwned && p.iconHandle != 0 {
		_, _, _ = procDestroyIcon.Call(uintptr(p.iconHandle))
	}
	p.iconHandle = 0
	p.iconOwned = false
}

func (p *Systray) destroyWindowClass() {
	if p.notifyClassName == nil || p.hinst == 0 {
		return
	}
	_, _, _ = procUnregisterClassW.Call(
		uintptr(unsafe.Pointer(p.notifyClassName)),
		uintptr(p.hinst),
	)
}

// PostThreadMessage posts a message to another thread's queue (typically WMTrayShutdown).
func PostThreadMessage(threadID uint32, msg uint32, wParam, lParam uintptr) error {
	ret, _, _ := procPostThreadMessage.Call(uintptr(threadID), uintptr(msg), wParam, lParam)
	if ret == 0 {
		return errors.New("PostThreadMessage failed")
	}
	return nil
}

// Quit ends the Run message loop. Must be called on the tray thread; other goroutines
// should use PostThreadMessage with WMTrayShutdown instead.
func (p *Systray) Quit() {
	procPostQuitMessage.Call(0)
}

// Show shows custom tray icon.
func (p *Systray) Show(iconResID uint16, hint string) error {
	icon, _, _ := LoadIcon.Call(uintptr(p.hinst), uintptr(iconResID))
	if icon == 0 {
		icon, _, _ = LoadIcon.Call(0, uintptr(IDI_APPLICATION))
	}

	err := p.setIcon(HICON(icon), false)
	if err != nil {
		return err
	}
	err = p.SetTooltip(hint)
	if err != nil {
		return err
	}
	return p.SetVisible(true)
}

func loadIconFile(file string) (HICON, bool, error) {
	path, err := filepath.Abs(file)
	if err != nil {
		return 0, false, err
	}
	icon, err := NewIconFromFile(path)
	if err != nil {
		return 0, false, err
	}
	return HICON(icon), true, nil
}

func copyUTF16Field(dst []uint16, src string) {
	u := syscall.StringToUTF16(src)
	if len(u) > len(dst) {
		u = u[:len(dst)]
		if len(u) > 0 {
			u[len(u)-1] = 0
		}
	}
	copy(dst, u)
}

// ShowCustom shows custom tray icon.
func (p *Systray) ShowCustom(file string, hint string) error {
	hicon, owned, err := loadIconFile(file)
	if err != nil {
		log.Printf("w32systray: load icon %q failed, falling back to default: %v", file, err)
		icon, _, _ := LoadIcon.Call(0, uintptr(IDI_APPLICATION))
		hicon, owned = HICON(icon), false
	}

	err = p.setIcon(hicon, owned)
	if err != nil {
		return err
	}
	err = p.SetTooltip(hint)
	if err != nil {
		return err
	}
	return p.SetVisible(true)
}

func (p *Systray) OnClick(fn func()) {
	if fn == nil {
		p.lclick = func() {}
		return
	}
	p.lclick = fn
}

func (p *Systray) OnRightClick(fn func()) {
	if fn == nil {
		p.rclick = func() {}
		return
	}
	p.rclick = fn
}

func (p *Systray) SetTooltip(tooltip string) error {
	nid := NOTIFYICONDATA{
		UID:  p.id,
		HWnd: HWND(p.hwnd),
	}
	nid.CbSize = uint32(unsafe.Sizeof(nid))

	nid.UFlags = NIF_TIP
	copyUTF16Field(nid.SzTip[:], tooltip)

	ret, _, _ := Shell_NotifyIcon.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&nid)))
	if ret == 0 {
		return errors.New("shell notify tooltip failed")
	}
	p.tooltip = tooltip
	return nil
}

func (p *Systray) ShowMessage(title, msg string, bigIcon bool) error {
	nid := NOTIFYICONDATA{
		UID:  p.id,
		HWnd: HWND(p.hwnd),
	}
	if bigIcon == true {
		nid.DwInfoFlags = NIIF_USER
	}

	nid.CbSize = uint32(unsafe.Sizeof(nid))

	nid.UFlags = NIF_INFO
	copyUTF16Field(nid.SzInfoTitle[:], title)
	copyUTF16Field(nid.SzInfo[:], msg)

	ret, _, _ := Shell_NotifyIcon.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&nid)))
	if ret == 0 {
		return errors.New("shell notify balloon failed")
	}
	return nil
}

func (p *Systray) SetVisible(visible bool) error {
	nid := NOTIFYICONDATA{
		UID:  p.id,
		HWnd: HWND(p.hwnd),
	}
	nid.CbSize = uint32(unsafe.Sizeof(nid))

	nid.UFlags = NIF_STATE
	nid.DwStateMask = NIS_HIDDEN
	if !visible {
		nid.DwState = NIS_HIDDEN
	}

	ret, _, _ := Shell_NotifyIcon.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&nid)))
	if ret == 0 {
		return errors.New("shell notify visibility failed")
	}
	p.visible = visible
	return nil
}

func (p *Systray) SetIcon(hicon HICON) error {
	return p.setIcon(hicon, false)
}

func (p *Systray) setIcon(hicon HICON, owned bool) error {
	if hicon != 0 {
		if p.iconHandle != 0 && p.iconHandle != hicon && p.iconOwned {
			_, _, _ = procDestroyIcon.Call(uintptr(p.iconHandle))
		}
		if hicon == p.iconHandle && p.iconOwned {
			owned = true
		}
	}
	// hicon==0 clears the shell icon only; p.iconHandle is kept for TaskbarCreated re-register.
	nid := NOTIFYICONDATA{
		UID:  p.id,
		HWnd: HWND(p.hwnd),
	}
	nid.CbSize = uint32(unsafe.Sizeof(nid))

	nid.UFlags = NIF_ICON
	if hicon == 0 {
		nid.HIcon = 0
	} else {
		nid.HIcon = hicon
	}

	ret, _, _ := Shell_NotifyIcon.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&nid)))
	if ret == 0 {
		return errors.New("shell notify icon failed")
	}
	if hicon != 0 {
		p.iconHandle = hicon
		p.iconOwned = owned
	}
	return nil
}

func (p *Systray) WinProc(hwnd HWND, msg uint32, wparam, lparam uintptr) uintptr {
	switch msg {
	case p.wmTaskbarCreated:
		if p.wmTaskbarCreated != 0 {
			if err := p.reregisterNotifyIcon(); err != nil {
				log.Printf("w32systray: TaskbarCreated reregister: %v", err)
			}
		}
	case NotifyIconMessageId:
		if lparam == WM_LBUTTONUP {
			if p.lclick != nil {
				p.lclick()
			}
			if len(p.Menu) > 0 {
				p.displayMenu(p.Menu)
			}
		} else if lparam == WM_RBUTTONUP {
			if p.rclick != nil {
				p.rclick()
			}
			if len(p.Menu) > 0 {
				p.displayMenu(p.Menu)
			}
		}

	case WM_COMMAND:
		cmdMsgID := int(wparam & 0xffff)
		switch cmdMsgID {
		default:
			if cmdMsgID >= MenuItemMsgID && cmdMsgID < (MenuItemMsgID+len(p.Menu)) {
				itemIndex := cmdMsgID - MenuItemMsgID
				menuItem := p.Menu[itemIndex]
				if menuItem.OnClick != nil {
					menuItem.OnClick()
				}
			}
		}
	}

	result, _, _ := DefWindowProc.Call(uintptr(hwnd), uintptr(msg), wparam, lparam)
	return result
}

func (p *Systray) reregisterNotifyIcon() error {
	nid := NOTIFYICONDATA{
		HWnd:             HWND(p.hwnd),
		UFlags:           NIF_MESSAGE | NIF_STATE,
		DwState:          NIS_HIDDEN,
		DwStateMask:      NIS_HIDDEN,
		UCallbackMessage: NotifyIconMessageId,
	}
	nid.CbSize = uint32(unsafe.Sizeof(nid))

	ret, _, _ := Shell_NotifyIcon.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
	if ret == 0 {
		return errors.New("shell notify recreate failed")
	}
	p.id = nid.UID

	nid.UVersion = NOTIFYICON_VERSION
	ret, _, _ = Shell_NotifyIcon.Call(NIM_SETVERSION, uintptr(unsafe.Pointer(&nid)))
	if ret == 0 {
		return errors.New("shell notify version failed")
	}

	if p.iconHandle != 0 {
		if err := p.setIcon(p.iconHandle, p.iconOwned); err != nil {
			return err
		}
	}
	if p.tooltip != "" {
		if err := p.SetTooltip(p.tooltip); err != nil {
			return err
		}
	}
	return p.SetVisible(p.visible)
}

func (p *Systray) Run() error {
	hwnd := p.hwnd
	var msg MSG
	for {
		rt, _, _ := GetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		switch int(rt) {
		case 0:
			return nil
		case -1:
			return errors.New("run failed")
		}

		if msg.Message == WMTrayShutdown {
			_ = p.Destroy()
			return nil
		}

		is, _, _ := IsDialogMessage.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&msg)))
		if is == 0 {
			TranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			DispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
		}
	}
}

func (p *Systray) displayMenu(menuItems []*MenuItem) error {
	ret, _, _ := CreatePopupMenu.Call(0, 0, 0, 0)
	if ret == 0 {
		return errors.New("can not create menu")
	}
	menu := HMENU(ret)
	defer func() {
		_, _, _ = procDestroyMenu.Call(uintptr(menu))
	}()

	for index, item := range menuItems {
		var ret bool
		itemID := MenuItemMsgID + index
		flags := MF_STRING
		if item.Disabled {
			flags = flags | MF_GRAYED
		}
		if item.Checked {
			flags = flags | MF_CHECKED
		}
		if item.BarBreak {
			flags = flags | MF_MENUBARBREAK
		}
		if item.Separator {
			flags = flags | MF_SEPARATOR
		}

		ret = appendMenu(menu, uintptr(flags), uintptr(itemID), item.Label)
		if ret == false {
			return errors.New("AppendMenu failed")
		}
	}

	x, y, ok := getCursorPos()
	if ok == false {
		return errors.New("GetCursorPos failed")
	}

	if setForegroundWindow(p.hwnd) == false {
		return errors.New("SetForegroundWindow failed")
	}

	if trackPopupMenu(menu, TPM_LEFTALIGN, x, y-5, p.hwnd) == false {
		return errors.New("TrackPopupMenu failed")
	}

	if ret, _, _ := procPostMessage.Call(uintptr(p.hwnd), uintptr(WM_NULL), 0, 0); ret == 0 {
		return errors.New("PostMessage failed")
	}
	return nil
}

func trackPopupMenu(menu HMENU, flags uint, x, y int, wnd HWND) bool {
	ret, _, _ := procTrackPopupMenu.Call(
		uintptr(menu),
		uintptr(flags),
		uintptr(x),
		uintptr(y),
		0,
		uintptr(wnd),
		0,
	)
	return ret != 0
}

func setForegroundWindow(wnd HWND) bool {
	ret, _, _ := procSetForegroundWindow.Call(
		uintptr(wnd),
	)
	return ret != 0
}

func getCursorPos() (x, y int, ok bool) {
	pt := POINT{}
	ret, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	return int(pt.X), int(pt.Y), ret != 0
}

func appendMenu(menu HMENU, flags uintptr, id uintptr, text string) bool {
	ret, _, _ := procAppendMenuW.Call(
		uintptr(menu),
		flags,
		id,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
	)
	return ret != 0
}

func NewIconFromFile(filePath string) (uintptr, error) {
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return 0, err
	}
	hicon, _, _ := LoadImage.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(absFilePath))),
		IMAGE_ICON,
		0,
		0,
		LR_DEFAULTSIZE|LR_LOADFROMFILE)
	if hicon == 0 {
		return 0, errors.New("load image failed: " + filePath)
	}
	return hicon, nil
}

func RegisterWindow(name string, proc WindowProc) (HINSTANCE, error) {
	hinst, _, _ := GetModuleHandle.Call(0)
	if hinst == 0 {
		return 0, errors.New("get module handle failed")
	}
	hicon, _, _ := LoadIcon.Call(0, uintptr(IDI_APPLICATION))
	if hicon == 0 {
		return 0, errors.New("load icon failed")
	}
	hcursor, _, _ := LoadCursor.Call(0, uintptr(IDC_ARROW))
	if hcursor == 0 {
		return 0, errors.New("load cursor failed")
	}

	hi := HINSTANCE(hinst)

	var wc WNDCLASSEX
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	wc.LpfnWndProc = syscall.NewCallback(proc)
	wc.HInstance = hi
	wc.HIcon = HICON(hicon)
	wc.HCursor = HCURSOR(hcursor)
	wc.HbrBackground = COLOR_BTNFACE + 1
	wc.LpszClassName = syscall.StringToUTF16Ptr(name)

	atom, _, _ := RegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 {
		return 0, errors.New("register class failed")
	}
	return hi, nil
}
