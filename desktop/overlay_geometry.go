package main

import (
	"context"
	"fmt"
	"math"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	overlayDockWidth        = 460
	overlayDockHeight       = 184
	overlayDockMarginX      = 24
	overlayDockMarginTop    = 18
	overlayDockMarginBottom = 56
)

type overlayRect struct {
	X      int
	Y      int
	Width  int
	Height int
}

type overlayDisplayBounds struct {
	X      int
	Y      int
	Width  int
	Height int

	OriginX int
	OriginY int

	MonitorWidth  int
	MonitorHeight int
	IsPrimary     bool
	IsCurrent     bool
}

type overlayLayout struct {
	Stage     overlayRect
	Dock      overlayRect
	Reference overlayDisplayBounds
}

func overlayLayoutForDisplays(displays []overlayDisplayBounds) (overlayLayout, error) {
	if len(displays) == 0 {
		return overlayLayout{}, fmt.Errorf("no overlay displays")
	}

	valid := make([]overlayDisplayBounds, 0, len(displays))
	for _, display := range displays {
		if display.Width <= 0 || display.Height <= 0 {
			continue
		}
		if display.MonitorWidth <= 0 {
			display.MonitorWidth = display.Width
		}
		if display.MonitorHeight <= 0 {
			display.MonitorHeight = display.Height
		}
		valid = append(valid, display)
	}
	if len(valid) == 0 {
		return overlayLayout{}, fmt.Errorf("no usable overlay displays")
	}

	minX := valid[0].X
	minY := valid[0].Y
	maxX := valid[0].X + valid[0].Width
	maxY := valid[0].Y + valid[0].Height
	for _, display := range valid[1:] {
		minX = min(minX, display.X)
		minY = min(minY, display.Y)
		maxX = max(maxX, display.X+display.Width)
		maxY = max(maxY, display.Y+display.Height)
	}

	primary := selectPrimaryDisplay(valid)
	return overlayLayout{
		Stage: overlayRect{
			X:      minX,
			Y:      minY,
			Width:  max(maxX-minX, overlayDockWidth),
			Height: max(maxY-minY, overlayDockHeight),
		},
		Dock:      dockRectForDisplay(primary),
		Reference: selectReferenceDisplay(valid),
	}, nil
}

func overlayDisplaysForRuntime(ctx context.Context) ([]overlayDisplayBounds, error) {
	displays, err := systemOverlayDisplays()
	if err != nil || len(displays) == 0 {
		return fallbackOverlayDisplays(ctx), err
	}
	return applyCurrentDisplayHint(ctx, displays), nil
}

func fallbackOverlayDisplays(ctx context.Context) []overlayDisplayBounds {
	screens, err := wailsruntime.ScreenGetAll(ctx)
	if err != nil || len(screens) == 0 {
		return []overlayDisplayBounds{{
			Width:         overlayDockWidth + overlayDockMarginX*2,
			Height:        overlayDockHeight + overlayDockMarginBottom + overlayDockMarginTop,
			MonitorWidth:  overlayDockWidth + overlayDockMarginX*2,
			MonitorHeight: overlayDockHeight + overlayDockMarginBottom + overlayDockMarginTop,
			IsPrimary:     true,
			IsCurrent:     true,
		}}
	}

	displays := make([]overlayDisplayBounds, 0, len(screens))
	nextX := 0
	for _, screen := range screens {
		width, height := screenDimensions(screen.Width, screen.Height, screen.Size.Width, screen.Size.Height)
		display := overlayDisplayBounds{
			X:             nextX,
			OriginX:       nextX,
			Width:         width,
			Height:        height,
			MonitorWidth:  width,
			MonitorHeight: height,
			IsPrimary:     screen.IsPrimary,
			IsCurrent:     screen.IsCurrent,
		}
		if screen.IsPrimary {
			display.X = 0
			display.OriginX = 0
		}
		displays = append(displays, display)
		nextX += width
	}
	return displays
}

func applyCurrentDisplayHint(ctx context.Context, displays []overlayDisplayBounds) []overlayDisplayBounds {
	screens, err := wailsruntime.ScreenGetAll(ctx)
	if err != nil || len(screens) == 0 {
		return displays
	}

	currentIndex := -1
	for i, screen := range screens {
		if screen.IsCurrent {
			currentIndex = i
			break
		}
	}
	if currentIndex < 0 {
		return displays
	}

	current := screens[currentIndex]
	currentWidth, currentHeight := screenDimensions(
		current.Width,
		current.Height,
		current.Size.Width,
		current.Size.Height,
	)

	next := make([]overlayDisplayBounds, len(displays))
	copy(next, displays)
	for i := range next {
		next[i].IsCurrent = false
	}

	if current.IsPrimary {
		for i := range next {
			if next[i].IsPrimary {
				next[i].IsCurrent = true
				return next
			}
		}
	}

	match := -1
	for i, display := range next {
		if display.MonitorWidth == currentWidth && display.MonitorHeight == currentHeight {
			if match == -1 || !display.IsPrimary {
				match = i
			}
		}
	}
	if match >= 0 {
		next[match].IsCurrent = true
	}
	return next
}

func screenDimensions(width int, height int, logicalWidth int, logicalHeight int) (int, int) {
	if logicalWidth > 0 && logicalHeight > 0 {
		return logicalWidth, logicalHeight
	}
	return width, height
}

func selectPrimaryDisplay(displays []overlayDisplayBounds) overlayDisplayBounds {
	for _, display := range displays {
		if display.IsPrimary {
			return display
		}
	}
	return displays[0]
}

func selectReferenceDisplay(displays []overlayDisplayBounds) overlayDisplayBounds {
	for _, display := range displays {
		if display.IsCurrent {
			return display
		}
	}
	return selectPrimaryDisplay(displays)
}

func dockRectForDisplay(display overlayDisplayBounds) overlayRect {
	return overlayRect{
		X:      dockCoordinate(display.X, display.Width, overlayDockWidth, overlayDockMarginX, overlayDockMarginX),
		Y:      dockCoordinate(display.Y, display.Height, overlayDockHeight, overlayDockMarginTop, overlayDockMarginBottom),
		Width:  overlayDockWidth,
		Height: overlayDockHeight,
	}
}

func dockCoordinate(start int, length int, size int, leadingMargin int, trailingMargin int) int {
	minPosition := start + leadingMargin
	maxPosition := start + length - size - trailingMargin
	if maxPosition >= minPosition {
		return maxPosition
	}
	if length > size {
		return start + int(math.Round(float64(length-size)/2))
	}
	return start
}

func overlayPositionForRect(rect overlayRect, reference overlayDisplayBounds) (int, int) {
	return rect.X - reference.OriginX, rect.Y - reference.OriginY
}
