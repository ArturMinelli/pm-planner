package reminder

import (
	"testing"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/config"
)

func TestBuildDayPlanNoRecordsUsesAnchors(t *testing.T) {
	date := testDate(t)
	anchors := []string{"08:00", "12:00", "13:30", "18:00"}

	day, err := BuildDayPlan(date, &api.WorkDay{}, anchors, date)
	if err != nil {
		t.Fatal(err)
	}
	if len(day.Slots) != 4 {
		t.Fatalf("slots: got %d", len(day.Slots))
	}
	for i, want := range []string{"08:00", "12:00", "13:30", "18:00"} {
		if day.Slots[i].Completed {
			t.Fatalf("slot %d should not be completed", i)
		}
		if got := day.Slots[i].Time.Format("15:04"); got != want {
			t.Fatalf("slot %d time: got %q want %q", i, got, want)
		}
	}
}

func TestBuildScheduleSuppressesCompletedSlots(t *testing.T) {
	date := testDate(t)
	anchors := []string{"08:00", "12:00", "13:30", "18:00"}
	wd := &api.WorkDay{ShiftTime: float64((8*time.Hour + 30*time.Minute).Seconds())}
	wd.TimeCards = append(wd.TimeCards,
		struct {
			Time string `json:"time"`
		}{Time: "08:00"},
		struct {
			Time string `json:"time"`
		}{Time: "12:00"},
		struct {
			Time string `json:"time"`
		}{Time: "13:30"},
	)

	day, err := BuildDayPlan(date, wd, anchors, date)
	if err != nil {
		t.Fatal(err)
	}
	settings, err := config.ResolveReminders(&config.File{
		Reminders: &config.Reminders{Enabled: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	scheduled := BuildSchedule(day, settings)
	if len(scheduled) != 2 {
		t.Fatalf("scheduled: got %d %#v", len(scheduled), scheduled)
	}
	wantSlotKey := SlotKey(1, true)
	for _, item := range scheduled {
		if item.SlotKey != wantSlotKey {
			t.Fatalf("only %q should remain, got %q", wantSlotKey, item.SlotKey)
		}
	}
}

func TestSlotKeyAndLabel(t *testing.T) {
	tests := []struct {
		journeyIndex int
		isExit       bool
		wantKey      string
		wantLabel    string
	}{
		{0, false, "in1", "Entrada 1"},
		{0, true, "out1", "Saída 1"},
		{1, false, "in2", "Entrada 2"},
		{1, true, "out2", "Saída 2"},
		{2, false, "in3", "Entrada 3"},
		{2, true, "out3", "Saída 3"},
	}
	for _, testCase := range tests {
		if got := SlotKey(testCase.journeyIndex, testCase.isExit); got != testCase.wantKey {
			t.Errorf("SlotKey(%d, %v) = %q, want %q", testCase.journeyIndex, testCase.isExit, got, testCase.wantKey)
		}
		if got := SlotLabel(testCase.journeyIndex, testCase.isExit); got != testCase.wantLabel {
			t.Errorf("SlotLabel(%d, %v) = %q, want %q", testCase.journeyIndex, testCase.isExit, got, testCase.wantLabel)
		}
	}
}

func TestBuildDayPlanTwoJourneys(t *testing.T) {
	date := testDate(t)
	anchors := []string{"08:00", "12:00", "13:30", "18:00"}
	wd := &api.WorkDay{ShiftTime: float64((8*time.Hour + 30*time.Minute).Seconds())}
	wd.TimeCards = append(wd.TimeCards,
		struct {
			Time string `json:"time"`
		}{Time: "08:00"},
		struct {
			Time string `json:"time"`
		}{Time: "12:00"},
		struct {
			Time string `json:"time"`
		}{Time: "13:30"},
	)

	dayPlan, err := BuildDayPlan(date, wd, anchors, date)
	if err != nil {
		t.Fatal(err)
	}
	if len(dayPlan.Slots) != 4 {
		t.Fatalf("expected 4 slots for 2 journeys, got %d", len(dayPlan.Slots))
	}

	wantKeys := []string{"in1", "out1", "in2", "out2"}
	wantLabels := []string{"Entrada 1", "Saída 1", "Entrada 2", "Saída 2"}
	for slotIndex, slot := range dayPlan.Slots {
		if slot.Key != wantKeys[slotIndex] {
			t.Errorf("slot[%d].Key = %q, want %q", slotIndex, slot.Key, wantKeys[slotIndex])
		}
		if slot.Label != wantLabels[slotIndex] {
			t.Errorf("slot[%d].Label = %q, want %q", slotIndex, slot.Label, wantLabels[slotIndex])
		}
	}
}

func TestBuildDayPlanThreeJourneys(t *testing.T) {
	date := testDate(t)
	anchors := []string{"08:00", "12:00", "13:30", "18:00"}
	wd := &api.WorkDay{ShiftTime: float64((8*time.Hour + 30*time.Minute).Seconds())}
	wd.TimeCards = append(wd.TimeCards,
		struct {
			Time string `json:"time"`
		}{Time: "08:00"},
		struct {
			Time string `json:"time"`
		}{Time: "11:00"},
		struct {
			Time string `json:"time"`
		}{Time: "12:00"},
		struct {
			Time string `json:"time"`
		}{Time: "13:30"},
		struct {
			Time string `json:"time"`
		}{Time: "15:00"},
		struct {
			Time string `json:"time"`
		}{Time: "17:00"},
	)

	dayPlan, err := BuildDayPlan(date, wd, anchors, date)
	if err != nil {
		t.Fatal(err)
	}
	if len(dayPlan.Slots) != 6 {
		t.Fatalf("expected 6 slots for 3 journeys, got %d", len(dayPlan.Slots))
	}

	wantKeys := []string{"in1", "out1", "in2", "out2", "in3", "out3"}
	wantLabels := []string{"Entrada 1", "Saída 1", "Entrada 2", "Saída 2", "Entrada 3", "Saída 3"}
	for slotIndex, slot := range dayPlan.Slots {
		if slot.Key != wantKeys[slotIndex] {
			t.Errorf("slot[%d].Key = %q, want %q", slotIndex, slot.Key, wantKeys[slotIndex])
		}
		if slot.Label != wantLabels[slotIndex] {
			t.Errorf("slot[%d].Label = %q, want %q", slotIndex, slot.Label, wantLabels[slotIndex])
		}
	}
}

func testDate(t *testing.T) time.Time {
	t.Helper()
	date, err := time.ParseInLocation("2006-01-02", "2026-06-08", time.Local)
	if err != nil {
		t.Fatal(err)
	}
	return date
}
