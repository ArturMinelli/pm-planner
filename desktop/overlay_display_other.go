//go:build !darwin && !linux && !windows

package main

import "fmt"

func systemOverlayDisplays() ([]overlayDisplayBounds, error) {
	return nil, fmt.Errorf("overlay display geometry is not supported on this platform")
}
