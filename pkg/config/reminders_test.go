package config

import (
	"strings"
	"testing"
)

func TestResolveReminders_defaults(t *testing.T) {
	got, err := ResolveReminders(&File{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.LeadMinutes) != 2 || got.LeadMinutes[0] != 15 || got.LeadMinutes[1] != 5 {
		t.Fatalf("lead minutes: %#v", got.LeadMinutes)
	}
	if !ReminderNativeEnabled(got) {
		t.Fatalf("native notifications should default enabled: %#v", got)
	}
}

func TestResolveReminders_sortsLeadMinutes(t *testing.T) {
	got, err := ResolveReminders(&File{
		Reminders: &Reminders{
			LeadMinutes: []int{5, 15},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.LeadMinutes[0] != 15 || got.LeadMinutes[1] != 5 {
		t.Fatalf("lead minutes: %#v", got.LeadMinutes)
	}
}

func TestResolveReminders_alwaysEnablesNativeNotifications(t *testing.T) {
	off := false
	got, err := ResolveReminders(&File{
		Reminders: &Reminders{
			LeadMinutes:         []int{15, 5},
			NativeNotifications: &off,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ReminderNativeEnabled(got) || got.NativeNotifications == nil || !*got.NativeNotifications {
		t.Fatalf("native notifications should always be enabled: %#v", got)
	}
}

func TestSave_rejectsInvalidReminders(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.yaml"
	err := Save(path, &File{
		Reminders: &Reminders{
			LeadMinutes: []int{15, 15},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate lead error, got %v", err)
	}
}
