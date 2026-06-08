//go:build !darwin && !linux && !windows

package main

func setOverlayInputPassthrough(bool) {}

func showOverlayWindowPassive() bool {
	return false
}
