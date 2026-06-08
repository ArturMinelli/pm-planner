package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"pm-cli/pkg/reminder"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	overlayWidth  = 430
	overlayHeight = 190
)

type OverlayApp struct {
	ctx     context.Context
	payload reminder.ScheduledReminder
}

func NewOverlayApp(payload reminder.ScheduledReminder) *OverlayApp {
	return &OverlayApp{payload: payload}
}

func (a *OverlayApp) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *OverlayApp) GetOverlayPayload() reminder.ScheduledReminder {
	return a.payload
}

func (a *OverlayApp) CloseOverlay() {
	if a.ctx != nil {
		wailsruntime.Quit(a.ctx)
	}
}

func runOverlay(encodedPayload string) error {
	payload, err := decodeOverlayPayload(encodedPayload)
	if err != nil {
		return err
	}
	overlayApp := NewOverlayApp(payload)
	return wails.Run(&options.App{
		Title:             "PM Planner Reminder",
		Width:             overlayWidth,
		Height:            overlayHeight,
		MinWidth:          overlayWidth,
		MinHeight:         overlayHeight,
		MaxWidth:          overlayWidth,
		MaxHeight:         overlayHeight,
		DisableResize:     true,
		Frameless:         true,
		AlwaysOnTop:       true,
		HideWindowOnClose: false,
		BackgroundColour:  options.NewRGBA(0, 0, 0, 0),
		AssetServer:       &assetserver.Options{Assets: assets},
		OnStartup:         overlayApp.Startup,
		OnDomReady:        positionOverlay,
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

func positionOverlay(ctx context.Context) {
	screens, err := wailsruntime.ScreenGetAll(ctx)
	if err != nil || len(screens) == 0 {
		wailsruntime.WindowCenter(ctx)
		return
	}
	screen := screens[0]
	for _, candidate := range screens {
		if candidate.IsCurrent || candidate.IsPrimary {
			screen = candidate
			break
		}
	}
	x := screen.Width - overlayWidth - 24
	y := screen.Height - overlayHeight - 56
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	wailsruntime.WindowSetPosition(ctx, x, y)
	wailsruntime.WindowSetAlwaysOnTop(ctx, true)
}
