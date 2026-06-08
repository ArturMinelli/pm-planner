package reminder

import (
	"testing"
	"time"

	"pm-cli/pkg/config"
)

type memoryStore struct {
	delivered map[string]bool
}

func (s *memoryStore) WasDelivered(id string) bool {
	return s.delivered[id]
}

func (s *memoryStore) MarkDelivered(id string, at time.Time) error {
	_ = at
	s.delivered[id] = true
	return nil
}

func TestBuildScheduleUsesLeadMinutes(t *testing.T) {
	date := testDate(t)
	day := &DayPlan{
		Date: date.Format("2006-01-02"),
		Slots: []Slot{
			{Key: SlotIn1, Label: "Entrada 1", Time: date.Add(8 * time.Hour)},
		},
	}
	settings, err := config.ResolveReminders(&config.File{
		Reminders: &config.Reminders{Enabled: true, LeadMinutes: []int{5, 15}},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := BuildSchedule(day, settings)
	if len(got) != 2 {
		t.Fatalf("schedule: got %d", len(got))
	}
	if got[0].LeadMinutes != 15 || got[0].FireAt.Format("15:04") != "07:45" {
		t.Fatalf("first reminder: %#v", got[0])
	}
	if got[1].LeadMinutes != 5 || got[1].FireAt.Format("15:04") != "07:55" {
		t.Fatalf("second reminder: %#v", got[1])
	}
}

func TestDueRemindersHonorsDeliveryWindowAndDedupe(t *testing.T) {
	date := testDate(t)
	day := &DayPlan{
		Date: date.Format("2006-01-02"),
		Slots: []Slot{
			{Key: SlotIn1, Label: "Entrada 1", Time: date.Add(8 * time.Hour)},
		},
	}
	settings, err := config.ResolveReminders(&config.File{
		Reminders: &config.Reminders{Enabled: true, LeadMinutes: []int{15}},
	})
	if err != nil {
		t.Fatal(err)
	}
	store := &memoryStore{delivered: map[string]bool{}}
	due := DueReminders(day, settings, store, nil, date.Add(7*time.Hour+45*time.Minute))
	if len(due) != 1 {
		t.Fatalf("due at fire time: got %d", len(due))
	}
	if err := store.MarkDelivered(due[0].ID, date); err != nil {
		t.Fatal(err)
	}
	due = DueReminders(day, settings, store, nil, date.Add(7*time.Hour+46*time.Minute))
	if len(due) != 0 {
		t.Fatalf("delivered reminder should be skipped: %#v", due)
	}
	due = DueReminders(day, settings, &memoryStore{delivered: map[string]bool{}}, nil, date.Add(8*time.Hour))
	if len(due) != 0 {
		t.Fatalf("stale reminder should be skipped: %#v", due)
	}
}

func TestNextWakeReturnsImmediateForDueReminder(t *testing.T) {
	date := testDate(t)
	day := &DayPlan{
		Date: date.Format("2006-01-02"),
		Slots: []Slot{
			{Key: SlotIn1, Label: "Entrada 1", Time: date.Add(8 * time.Hour)},
		},
	}
	settings, err := config.ResolveReminders(&config.File{
		Reminders: &config.Reminders{Enabled: true, LeadMinutes: []int{15}},
	})
	if err != nil {
		t.Fatal(err)
	}
	now := date.Add(7*time.Hour + 45*time.Minute)
	wake, ok := NextWake(day, settings, nil, nil, now)
	if !ok || !wake.Equal(now) {
		t.Fatalf("wake: got %v %v", wake, ok)
	}
}
