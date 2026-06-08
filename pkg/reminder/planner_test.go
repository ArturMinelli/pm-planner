package reminder

import (
	"testing"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/config"
)

func TestBuildDayPlanNoRecordsUsesAnchors(t *testing.T) {
	date := testDate(t)
	anchors := [4]string{"08:00", "12:00", "13:30", "18:00"}

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
	anchors := [4]string{"08:00", "12:00", "13:30", "18:00"}
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
	for _, item := range scheduled {
		if item.SlotKey != SlotOut2 {
			t.Fatalf("only out2 should remain, got %q", item.SlotKey)
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
