//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	overlayGWLExStyle      = ^uintptr(19)
	overlayWSExLayered     = 0x00080000
	overlayWSExNoActivate  = 0x08000000
	overlayWSExTransparent = 0x00000020

	overlaySWPNoSize       = 0x0001
	overlaySWPNoMove       = 0x0002
	overlaySWPNoZOrder     = 0x0004
	overlaySWPNoActivate   = 0x0010
	overlaySWPFrameChanged = 0x0020
	overlaySWPShowWindow   = 0x0040

	overlaySWShowNoActivate = 4
)

var (
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetWindowThreadProcessID = user32.NewProc("GetWindowThreadProcessId")
	procGetWindowTextLengthW     = user32.NewProc("GetWindowTextLengthW")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowLongPtrW        = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW        = user32.NewProc("SetWindowLongPtrW")
	procSetWindowPos             = user32.NewProc("SetWindowPos")
	procShowWindow               = user32.NewProc("ShowWindow")
)

type overlayWindowSearch struct {
	pid    uint32
	handle uintptr
}

func setOverlayInputPassthrough(enabled bool) {
	handle := currentOverlayWindowHandle()
	if handle == 0 {
		return
	}
	style, _, _ := procGetWindowLongPtrW.Call(handle, overlayGWLExStyle)
	next := style
	if enabled {
		next |= overlayWSExLayered | overlayWSExTransparent | overlayWSExNoActivate
	} else {
		next &^= overlayWSExTransparent | overlayWSExNoActivate
	}
	if next != style {
		procSetWindowLongPtrW.Call(handle, overlayGWLExStyle, next)
	}
	procSetWindowPos.Call(
		handle,
		0,
		0,
		0,
		0,
		0,
		overlaySWPNoMove|overlaySWPNoSize|overlaySWPNoZOrder|overlaySWPNoActivate|overlaySWPFrameChanged,
	)
}

func showOverlayWindowPassive() bool {
	handle := currentOverlayWindowHandle()
	if handle == 0 {
		return false
	}
	procShowWindow.Call(handle, overlaySWShowNoActivate)
	procSetWindowPos.Call(
		handle,
		^uintptr(0),
		0,
		0,
		0,
		0,
		overlaySWPNoMove|overlaySWPNoSize|overlaySWPNoActivate|overlaySWPShowWindow,
	)
	return true
}

func currentOverlayWindowHandle() uintptr {
	search := &overlayWindowSearch{pid: uint32(os.Getpid())}
	procEnumWindows.Call(syscall.NewCallback(enumOverlayWindow), uintptr(unsafe.Pointer(search)))
	return search.handle
}

func enumOverlayWindow(hwnd uintptr, data uintptr) uintptr {
	search := (*overlayWindowSearch)(unsafe.Pointer(data))
	var pid uint32
	procGetWindowThreadProcessID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid != search.pid {
		return 1
	}
	length, _, _ := procGetWindowTextLengthW.Call(hwnd)
	if length == 0 {
		return 1
	}
	buffer := make([]uint16, length+1)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buffer[0])), length+1)
	if syscall.UTF16ToString(buffer) != overlayWindowTitle {
		return 1
	}
	search.handle = hwnd
	return 0
}
