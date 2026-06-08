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
	if got.Enabled {
		t.Fatal("reminders should be opt-in")
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
			Enabled:     true,
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

func TestResolveRemindersClearsAutostartWhenDisabled(t *testing.T) {
	got, err := ResolveReminders(&File{
		Reminders: &Reminders{
			Enabled:     false,
			Autostart:   true,
			LeadMinutes: []int{15, 5},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Autostart {
		t.Fatal("autostart should be disabled when reminders are disabled")
	}
}

func TestResolveReminders_rejectsDisabledNativeNotificationsWhenEnabled(t *testing.T) {
	off := false
	_, err := ResolveReminders(&File{
		Reminders: &Reminders{
			Enabled:             true,
			LeadMinutes:         []int{15, 5},
			NativeNotifications: &off,
		},
	})
	if err == nil || !strings.Contains(err.Error(), "native notifications") {
		t.Fatalf("expected native notification validation error, got %v", err)
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
