//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

const monitorInfoPrimary = 1

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procEnumDisplayMonitors = user32.NewProc("EnumDisplayMonitors")
	procGetMonitorInfoW     = user32.NewProc("GetMonitorInfoW")
)

type windowsRect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type windowsMonitorInfo struct {
	Size    uint32
	Monitor windowsRect
	Work    windowsRect
	Flags   uint32
}

type windowsDisplayState struct {
	displays    []overlayDisplayBounds
	primaryWork windowsRect
	hasPrimary  bool
	err         error
}

func systemOverlayDisplays() ([]overlayDisplayBounds, error) {
	state := &windowsDisplayState{}
	callback := syscall.NewCallback(enumOverlayMonitor)
	ok, _, callErr := procEnumDisplayMonitors.Call(0, 0, callback, uintptr(unsafe.Pointer(state)))
	if ok == 0 {
		if callErr != syscall.Errno(0) {
			return nil, callErr
		}
		return nil, fmt.Errorf("EnumDisplayMonitors failed")
	}
	if state.err != nil {
		return nil, state.err
	}
	if len(state.displays) == 0 {
		return nil, fmt.Errorf("no Windows displays")
	}
	if !state.hasPrimary {
		state.primaryWork = windowsRect{}
		for i := range state.displays {
			state.displays[i].IsPrimary = i == 0
		}
	}
	for i := range state.displays {
		state.displays[i].X -= int(state.primaryWork.Left)
		state.displays[i].Y -= int(state.primaryWork.Top)
		state.displays[i].OriginX = state.displays[i].X
		state.displays[i].OriginY = state.displays[i].Y
	}
	return state.displays, nil
}

func enumOverlayMonitor(hMonitor uintptr, _ uintptr, _ uintptr, data uintptr) uintptr {
	state := (*windowsDisplayState)(unsafe.Pointer(data))
	info := windowsMonitorInfo{Size: uint32(unsafe.Sizeof(windowsMonitorInfo{}))}
	ok, _, callErr := procGetMonitorInfoW.Call(hMonitor, uintptr(unsafe.Pointer(&info)))
	if ok == 0 {
		if callErr != syscall.Errno(0) {
			state.err = callErr
		} else {
			state.err = fmt.Errorf("GetMonitorInfoW failed")
		}
		return 0
	}

	isPrimary := info.Flags&monitorInfoPrimary == monitorInfoPrimary
	if isPrimary {
		state.primaryWork = info.Work
		state.hasPrimary = true
	}
	state.displays = append(state.displays, overlayDisplayBounds{
		X:             int(info.Work.Left),
		Y:             int(info.Work.Top),
		Width:         int(info.Work.Right - info.Work.Left),
		Height:        int(info.Work.Bottom - info.Work.Top),
		MonitorWidth:  int(info.Monitor.Right - info.Monitor.Left),
		MonitorHeight: int(info.Monitor.Bottom - info.Monitor.Top),
		IsPrimary:     isPrimary,
	})
	return 1
}
