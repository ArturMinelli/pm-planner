//go:build linux

package main

/*
#cgo linux pkg-config: gtk+-3.0
#cgo CFLAGS: -w
#include "gtk/gtk.h"
#include "gdk/gdk.h"

typedef struct OverlayDisplay {
	int x;
	int y;
	int width;
	int height;
	int originX;
	int originY;
	int monitorWidth;
	int monitorHeight;
	int isPrimary;
} OverlayDisplay;

int OverlayDisplayCount() {
	GdkDisplay *display = gdk_display_get_default();
	if (display == NULL) {
		return 0;
	}
	return gdk_display_get_n_monitors(display);
}

OverlayDisplay OverlayDisplayAt(int nth) {
	GdkDisplay *display = gdk_display_get_default();
	GdkMonitor *primary = gdk_display_get_primary_monitor(display);
	GdkMonitor *monitor = gdk_display_get_monitor(display, nth);
	if (primary == NULL) {
		primary = gdk_display_get_monitor(display, 0);
	}

	GdkRectangle primaryGeometry;
	GdkRectangle geometry;
	GdkRectangle workarea;
	gdk_monitor_get_geometry(primary, &primaryGeometry);
	gdk_monitor_get_geometry(monitor, &geometry);
	gdk_monitor_get_workarea(monitor, &workarea);

	OverlayDisplay result;
	result.x = workarea.x - primaryGeometry.x;
	result.y = workarea.y - primaryGeometry.y;
	result.width = workarea.width;
	result.height = workarea.height;
	result.originX = geometry.x - primaryGeometry.x;
	result.originY = geometry.y - primaryGeometry.y;
	result.monitorWidth = geometry.width;
	result.monitorHeight = geometry.height;
	result.isPrimary = monitor == primary;
	return result;
}
*/
import "C"

import "fmt"

func systemOverlayDisplays() ([]overlayDisplayBounds, error) {
	count := int(C.OverlayDisplayCount())
	if count <= 0 {
		return nil, fmt.Errorf("no Linux displays")
	}
	displays := make([]overlayDisplayBounds, 0, count)
	for i := 0; i < count; i++ {
		display := C.OverlayDisplayAt(C.int(i))
		displays = append(displays, overlayDisplayBounds{
			X:             int(display.x),
			Y:             int(display.y),
			Width:         int(display.width),
			Height:        int(display.height),
			OriginX:       int(display.originX),
			OriginY:       int(display.originY),
			MonitorWidth:  int(display.monitorWidth),
			MonitorHeight: int(display.monitorHeight),
			IsPrimary:     display.isPrimary == 1,
		})
	}
	return displays, nil
}
