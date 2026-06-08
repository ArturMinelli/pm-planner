package main

import (
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
