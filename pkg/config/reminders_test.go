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
	if got.Animation != ReminderAnimationTrain {
		t.Fatalf("animation: got %q", got.Animation)
	}
	if len(got.LeadMinutes) != 2 || got.LeadMinutes[0] != 15 || got.LeadMinutes[1] != 5 {
		t.Fatalf("lead minutes: %#v", got.LeadMinutes)
	}
	if !ReminderNativeEnabled(got) || !ReminderPopupEnabled(got) {
		t.Fatalf("channels should default enabled: %#v", got)
	}
}

func TestResolveReminders_sortsLeadMinutes(t *testing.T) {
	got, err := ResolveReminders(&File{
		Reminders: &Reminders{
			Enabled:     true,
			LeadMinutes: []int{5, 15},
			Animation:   ReminderAnimationRocket,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.LeadMinutes[0] != 15 || got.LeadMinutes[1] != 5 {
		t.Fatalf("lead minutes: %#v", got.LeadMinutes)
	}
	if got.Animation != ReminderAnimationRocket {
		t.Fatalf("animation: %q", got.Animation)
	}
}

func TestResolveRemindersClearsAutostartWhenDisabled(t *testing.T) {
	got, err := ResolveReminders(&File{
		Reminders: &Reminders{
			Enabled:     false,
			Autostart:   true,
			LeadMinutes: []int{15, 5},
			Animation:   ReminderAnimationTrain,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Autostart {
		t.Fatal("autostart should be disabled when reminders are disabled")
	}
}

func TestResolveReminders_rejectsDisabledChannelsWhenEnabled(t *testing.T) {
	off := false
	_, err := ResolveReminders(&File{
		Reminders: &Reminders{
			Enabled:             true,
			LeadMinutes:         []int{15, 5},
			Animation:           ReminderAnimationTrain,
			NativeNotifications: &off,
			PopupNotifications:  &off,
		},
	})
	if err == nil || !strings.Contains(err.Error(), "notification channel") {
		t.Fatalf("expected channel validation error, got %v", err)
	}
}

func TestSave_rejectsInvalidReminders(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.yaml"
	err := Save(path, &File{
		Reminders: &Reminders{
			LeadMinutes: []int{15, 15},
			Animation:   ReminderAnimationTrain,
		},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate lead error, got %v", err)
	}
}
