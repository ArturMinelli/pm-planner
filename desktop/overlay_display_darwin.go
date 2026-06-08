//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit
#import <AppKit/AppKit.h>

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
	return (int)[[NSScreen screens] count];
}

OverlayDisplay OverlayDisplayAt(int nth) {
	NSArray<NSScreen *> *screens = [NSScreen screens];
	NSScreen *primary = [screens objectAtIndex:0];
	NSScreen *screen = [screens objectAtIndex:nth];
	NSRect primaryVisible = [primary visibleFrame];
	NSRect visible = [screen visibleFrame];
	NSRect frame = [screen frame];
	CGFloat primaryTop = primaryVisible.origin.y + primaryVisible.size.height;

	OverlayDisplay display;
	display.x = (int)(visible.origin.x - primaryVisible.origin.x);
	display.y = (int)(primaryTop - (visible.origin.y + visible.size.height));
	display.width = (int)visible.size.width;
	display.height = (int)visible.size.height;
	display.originX = display.x;
	display.originY = display.y;
	display.monitorWidth = (int)frame.size.width;
	display.monitorHeight = (int)frame.size.height;
	display.isPrimary = nth == 0;
	return display;
}
*/
import "C"

import "fmt"

func systemOverlayDisplays() ([]overlayDisplayBounds, error) {
	count := int(C.OverlayDisplayCount())
	if count <= 0 {
		return nil, fmt.Errorf("no macOS displays")
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
