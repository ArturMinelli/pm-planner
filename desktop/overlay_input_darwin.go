//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit
#import <AppKit/AppKit.h>
#import <dispatch/dispatch.h>

void PMSetOverlayInputPassthrough(int enabled) {
	dispatch_async(dispatch_get_main_queue(), ^{
		for (NSWindow *window in [[NSApplication sharedApplication] windows]) {
			[window setIgnoresMouseEvents:enabled ? YES : NO];
		}
	});
}

int PMShowOverlayWindowPassive() {
	dispatch_async(dispatch_get_main_queue(), ^{
		for (NSWindow *window in [[NSApplication sharedApplication] windows]) {
			[window setIgnoresMouseEvents:YES];
			[window orderFrontRegardless];
		}
	});
	return 1;
}
*/
import "C"

func setOverlayInputPassthrough(enabled bool) {
	if enabled {
		C.PMSetOverlayInputPassthrough(1)
		return
	}
	C.PMSetOverlayInputPassthrough(0)
}

func showOverlayWindowPassive() bool {
	return C.PMShowOverlayWindowPassive() == 1
}
