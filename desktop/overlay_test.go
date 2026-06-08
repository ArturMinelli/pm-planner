package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"pm-cli/pkg/reminder"
)

func TestDecodeOverlayPayload(t *testing.T) {
	event := reminder.ScheduledReminder{
		ID:          "2026-06-08|out2|5",
		Date:        "2026-06-08",
		SlotKey:     reminder.SlotOut2,
		SlotLabel:   "Saída 2",
		SlotTime:    time.Date(2026, 6, 8, 18, 0, 0, 0, time.Local),
		LeadMinutes: 5,
		FireAt:      time.Date(2026, 6, 8, 17, 55, 0, 0, time.Local),
		Animation:   "train",
	}
	b, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	got, err := decodeOverlayPayload(base64.StdEncoding.EncodeToString(b))
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != event.ID || got.SlotLabel != event.SlotLabel || got.Animation != event.Animation {
		t.Fatalf("decoded payload: %#v", got)
	}
}

func TestDockOverlayWithoutContextIsNoop(t *testing.T) {
	window := &fakeOverlayWindow{}
	app := NewOverlayApp(reminder.ScheduledReminder{})
	app.window = window

	app.DockOverlay()
	app.DockOverlay()

	if window.setRectCalls != 0 {
		t.Fatalf("set rect calls: got %d, want 0", window.setRectCalls)
	}
}

func TestDockOverlayIsIdempotent(t *testing.T) {
	window := &fakeOverlayWindow{}
	app := NewOverlayApp(reminder.ScheduledReminder{})
	app.window = window
	app.ctx = context.Background()
	app.hasLayout = true
	app.layout = overlayLayout{
		Dock:      overlayRect{X: 100, Y: 120, Width: overlayDockWidth, Height: overlayDockHeight},
		Reference: overlayDisplayBounds{OriginX: 20, OriginY: 40},
	}

	app.DockOverlay()
	app.DockOverlay()

	if window.setRectCalls != 1 {
		t.Fatalf("set rect calls: got %d, want 1", window.setRectCalls)
	}
	if window.lastRect != app.layout.Dock {
		t.Fatalf("dock rect: got %#v, want %#v", window.lastRect, app.layout.Dock)
	}
	if !window.alwaysOnTop {
		t.Fatalf("always on top was not restored")
	}
	if len(window.inputPassthrough) != 1 || window.inputPassthrough[0] {
		t.Fatalf("input passthrough calls: got %#v, want [false]", window.inputPassthrough)
	}
}

func TestStageOverlayEnablesInputPassthroughAndPassiveShow(t *testing.T) {
	window := &fakeOverlayWindow{}
	layout := overlayLayout{
		Stage:     overlayRect{X: -100, Y: 20, Width: 1600, Height: 900},
		Dock:      overlayRect{X: 900, Y: 660, Width: overlayDockWidth, Height: overlayDockHeight},
		Reference: overlayDisplayBounds{OriginX: -100, OriginY: 20},
	}
	app := NewOverlayApp(reminder.ScheduledReminder{})
	app.window = window
	app.build = func(context.Context) (overlayLayout, error) {
		return layout, nil
	}

	app.StageOverlay(context.Background())

	if len(window.inputPassthrough) != 1 || !window.inputPassthrough[0] {
		t.Fatalf("input passthrough calls: got %#v, want [true]", window.inputPassthrough)
	}
	if window.setRectCalls != 1 || window.lastRect != layout.Stage {
		t.Fatalf("stage rect calls: got %d %#v, want 1 %#v", window.setRectCalls, window.lastRect, layout.Stage)
	}
	if window.passiveShowCalls != 1 {
		t.Fatalf("passive show calls: got %d, want 1", window.passiveShowCalls)
	}
	if !window.alwaysOnTop {
		t.Fatalf("always on top was not set")
	}
}

func TestCloseOverlayRestoresInputPassthrough(t *testing.T) {
	window := &fakeOverlayWindow{}
	app := NewOverlayApp(reminder.ScheduledReminder{})
	app.window = window
	app.ctx = context.Background()

	app.CloseOverlay()

	if len(window.inputPassthrough) != 1 || window.inputPassthrough[0] {
		t.Fatalf("input passthrough calls: got %#v, want [false]", window.inputPassthrough)
	}
	if window.quitCalls != 1 {
		t.Fatalf("quit calls: got %d, want 1", window.quitCalls)
	}
}

type fakeOverlayWindow struct {
	setRectCalls     int
	passiveShowCalls int
	quitCalls        int
	lastRect         overlayRect
	alwaysOnTop      bool
	inputPassthrough []bool
}

func (w *fakeOverlayWindow) SetRect(_ context.Context, rect overlayRect, _ overlayDisplayBounds) {
	w.setRectCalls++
	w.lastRect = rect
}

func (w *fakeOverlayWindow) SetAlwaysOnTop(_ context.Context, value bool) {
	w.alwaysOnTop = value
}

func (w *fakeOverlayWindow) SetInputPassthrough(_ context.Context, value bool) {
	w.inputPassthrough = append(w.inputPassthrough, value)
}

func (w *fakeOverlayWindow) ShowPassive(context.Context) {
	w.passiveShowCalls++
}

func (w *fakeOverlayWindow) Show(context.Context) {}

func (w *fakeOverlayWindow) Center(context.Context) {}

func (w *fakeOverlayWindow) Quit(context.Context) {
	w.quitCalls++
}
