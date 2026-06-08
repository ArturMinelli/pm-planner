package main

import "testing"

func TestOverlayLayoutForSingleDisplay(t *testing.T) {
	layout, err := overlayLayoutForDisplays([]overlayDisplayBounds{{
		Width:     1440,
		Height:    900,
		IsPrimary: true,
		IsCurrent: true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	assertRect(t, layout.Stage, overlayRect{Width: 1440, Height: 900})
	assertRect(t, layout.Dock, overlayRect{
		X:      1440 - overlayDockWidth - overlayDockMarginX,
		Y:      900 - overlayDockHeight - overlayDockMarginBottom,
		Width:  overlayDockWidth,
		Height: overlayDockHeight,
	})
}

func TestOverlayLayoutForHorizontalDisplays(t *testing.T) {
	layout, err := overlayLayoutForDisplays([]overlayDisplayBounds{
		{Width: 1440, Height: 900, IsPrimary: true, IsCurrent: true},
		{X: 1440, OriginX: 1440, Width: 1920, Height: 1080},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertRect(t, layout.Stage, overlayRect{Width: 3360, Height: 1080})
	assertRect(t, layout.Dock, overlayRect{
		X:      1440 - overlayDockWidth - overlayDockMarginX,
		Y:      900 - overlayDockHeight - overlayDockMarginBottom,
		Width:  overlayDockWidth,
		Height: overlayDockHeight,
	})
}

func TestOverlayLayoutForNegativeLeftDisplay(t *testing.T) {
	layout, err := overlayLayoutForDisplays([]overlayDisplayBounds{
		{Width: 1440, Height: 900, IsPrimary: true},
		{X: -1920, OriginX: -1920, Width: 1920, Height: 1080, IsCurrent: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertRect(t, layout.Stage, overlayRect{X: -1920, Width: 3360, Height: 1080})
	x, y := overlayPositionForRect(layout.Stage, layout.Reference)
	if x != 0 || y != 0 {
		t.Fatalf("stage position from current left display: got %d,%d", x, y)
	}
}

func TestOverlayLayoutForVerticalDisplays(t *testing.T) {
	layout, err := overlayLayoutForDisplays([]overlayDisplayBounds{
		{Width: 1440, Height: 900, IsPrimary: true, IsCurrent: true},
		{Y: -900, OriginY: -900, Width: 1440, Height: 900},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertRect(t, layout.Stage, overlayRect{Y: -900, Width: 1440, Height: 1800})
}

func assertRect(t *testing.T, got overlayRect, want overlayRect) {
	t.Helper()
	if got != want {
		t.Fatalf("rect: got %#v, want %#v", got, want)
	}
}
