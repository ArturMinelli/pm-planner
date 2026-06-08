package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"pm-cli/pkg/reminder"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const overlayWindowTitle = "PM Planner Reminder"

type OverlayApp struct {
	mu        sync.Mutex
	ctx       context.Context
	payload   reminder.ScheduledReminder
	layout    overlayLayout
	hasLayout bool
	docked    bool
	window    overlayWindowController
	build     func(context.Context) (overlayLayout, error)
}

func NewOverlayApp(payload reminder.ScheduledReminder) *OverlayApp {
	return &OverlayApp{
		payload: payload,
		window:  wailsOverlayWindow{},
		build:   buildOverlayLayout,
	}
}

func (a *OverlayApp) Startup(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ctx = ctx
}

func (a *OverlayApp) GetOverlayPayload() reminder.ScheduledReminder {
	return a.payload
}

func (a *OverlayApp) CloseOverlay() {
	a.mu.Lock()
	ctx := a.ctx
	window := a.window
	a.mu.Unlock()
	if ctx != nil {
		window.SetInputPassthrough(ctx, false)
		window.Quit(ctx)
	}
}

func (a *OverlayApp) DockOverlay() {
	a.mu.Lock()
	if a.ctx == nil || a.docked {
		a.mu.Unlock()
		return
	}
	ctx := a.ctx
	layout := a.layout
	hasLayout := a.hasLayout
	window := a.window
	build := a.build
	a.docked = true
	a.mu.Unlock()

	window.SetInputPassthrough(ctx, false)
	if !hasLayout {
		next, err := build(ctx)
		if err != nil {
			return
		}
		layout = next
		a.mu.Lock()
		a.layout = layout
		a.hasLayout = true
		a.mu.Unlock()
	}

	window.SetRect(ctx, layout.Dock, layout.Reference)
	window.SetAlwaysOnTop(ctx, true)
}

func (a *OverlayApp) StageOverlay(ctx context.Context) {
	a.mu.Lock()
	a.ctx = ctx
	a.docked = false
	window := a.window
	build := a.build
	a.mu.Unlock()

	window.SetInputPassthrough(ctx, true)
	layout, err := build(ctx)
	if err != nil {
		window.Center(ctx)
		window.SetAlwaysOnTop(ctx, true)
		window.ShowPassive(ctx)
		return
	}

	a.mu.Lock()
	a.layout = layout
	a.hasLayout = true
	a.mu.Unlock()

	window.SetRect(ctx, layout.Stage, layout.Reference)
	window.SetAlwaysOnTop(ctx, true)
	window.ShowPassive(ctx)
}

func runOverlay(encodedPayload string) error {
	payload, err := decodeOverlayPayload(encodedPayload)
	if err != nil {
		return err
	}
	overlayApp := NewOverlayApp(payload)
	return wails.Run(&options.App{
		Title:             overlayWindowTitle,
		Width:             overlayDockWidth,
		Height:            overlayDockHeight,
		DisableResize:     true,
		Frameless:         true,
		StartHidden:       true,
		AlwaysOnTop:       true,
		HideWindowOnClose: false,
		BackgroundColour:  options.NewRGBA(0, 0, 0, 0),
		AssetServer:       &assetserver.Options{Assets: assets},
		OnStartup:         overlayApp.Startup,
		OnDomReady:        overlayApp.StageOverlay,
		Bind:              []interface{}{overlayApp},
		Windows:           &windows.Options{WindowIsTranslucent: true, WebviewIsTransparent: true, DisableFramelessWindowDecorations: true},
		Mac:               &mac.Options{WindowIsTranslucent: true, WebviewIsTransparent: true},
		Linux:             &linux.Options{WindowIsTranslucent: true, WebviewGpuPolicy: linux.WebviewGpuPolicyOnDemand},
	})
}

func decodeOverlayPayload(encoded string) (reminder.ScheduledReminder, error) {
	var payload reminder.ScheduledReminder
	if encoded == "" {
		return payload, fmt.Errorf("missing overlay payload")
	}
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return payload, err
	}
	if err := json.Unmarshal(b, &payload); err != nil {
		return payload, err
	}
	return payload, nil
}

func buildOverlayLayout(ctx context.Context) (overlayLayout, error) {
	displays, _ := overlayDisplaysForRuntime(ctx)
	layout, layoutErr := overlayLayoutForDisplays(displays)
	if layoutErr != nil {
		return overlayLayout{}, layoutErr
	}
	return layout, nil
}

type overlayWindowController interface {
	SetRect(context.Context, overlayRect, overlayDisplayBounds)
	SetAlwaysOnTop(context.Context, bool)
	SetInputPassthrough(context.Context, bool)
	ShowPassive(context.Context)
	Show(context.Context)
	Center(context.Context)
	Quit(context.Context)
}

type wailsOverlayWindow struct{}

func (wailsOverlayWindow) SetRect(ctx context.Context, rect overlayRect, reference overlayDisplayBounds) {
	x, y := overlayPositionForRect(rect, reference)
	wailsruntime.WindowSetPosition(ctx, x, y)
	wailsruntime.WindowSetSize(ctx, rect.Width, rect.Height)
}

func (wailsOverlayWindow) SetAlwaysOnTop(ctx context.Context, value bool) {
	wailsruntime.WindowSetAlwaysOnTop(ctx, value)
}

func (wailsOverlayWindow) SetInputPassthrough(_ context.Context, value bool) {
	setOverlayInputPassthrough(value)
}

func (wailsOverlayWindow) ShowPassive(ctx context.Context) {
	if !showOverlayWindowPassive() {
		wailsruntime.WindowShow(ctx)
	}
}

func (wailsOverlayWindow) Show(ctx context.Context) {
	wailsruntime.WindowShow(ctx)
}

func (wailsOverlayWindow) Center(ctx context.Context) {
	wailsruntime.WindowCenter(ctx)
}

func (wailsOverlayWindow) Quit(ctx context.Context) {
	wailsruntime.Quit(ctx)
}
