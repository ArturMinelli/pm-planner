package config

import (
	"fmt"
	"sort"
	"strings"
)

const (
	ReminderAnimationTrain  = "train"
	ReminderAnimationRocket = "rocket"
	ReminderAnimationBell   = "bell"
)

var defaultReminderLeads = []int{15, 5}

// Reminders holds user-configurable background reminder settings.
type Reminders struct {
	Enabled             bool   `json:"enabled" yaml:"enabled"`
	Autostart           bool   `json:"autostart" yaml:"autostart"`
	LeadMinutes         []int  `json:"lead_minutes,omitempty" yaml:"lead_minutes,omitempty"`
	Animation           string `json:"animation,omitempty" yaml:"animation,omitempty"`
	NativeNotifications *bool  `json:"native_notifications,omitempty" yaml:"native_notifications,omitempty"`
	PopupNotifications  *bool  `json:"popup_notifications,omitempty" yaml:"popup_notifications,omitempty"`
}

func boolPtr(v bool) *bool {
	return &v
}

// DefaultReminders returns the normalized reminder defaults. Reminders are opt-in.
func DefaultReminders() Reminders {
	return Reminders{
		Enabled:             false,
		Autostart:           false,
		LeadMinutes:         append([]int(nil), defaultReminderLeads...),
		Animation:           ReminderAnimationTrain,
		NativeNotifications: boolPtr(true),
		PopupNotifications:  boolPtr(true),
	}
}

// ResolveReminders merges config with defaults and validates the result.
func ResolveReminders(f *File) (Reminders, error) {
	out := DefaultReminders()
	if f != nil && f.Reminders != nil {
		raw := f.Reminders
		out.Enabled = raw.Enabled
		out.Autostart = raw.Autostart
		if len(raw.LeadMinutes) > 0 {
			out.LeadMinutes = append([]int(nil), raw.LeadMinutes...)
		}
		if strings.TrimSpace(raw.Animation) != "" {
			out.Animation = strings.TrimSpace(raw.Animation)
		}
		if raw.NativeNotifications != nil {
			out.NativeNotifications = boolPtr(*raw.NativeNotifications)
		}
		if raw.PopupNotifications != nil {
			out.PopupNotifications = boolPtr(*raw.PopupNotifications)
		}
	}
	normalizeReminderLeads(out.LeadMinutes)
	if !out.Enabled {
		out.Autostart = false
	}
	if err := ValidateReminders(out); err != nil {
		return Reminders{}, err
	}
	return out, nil
}

func normalizeReminderLeads(v []int) {
	sort.Slice(v, func(i, j int) bool { return v[i] > v[j] })
}

// ValidateReminders checks reminder settings are small, known, and non-empty when enabled.
func ValidateReminders(r Reminders) error {
	if len(r.LeadMinutes) == 0 {
		return fmt.Errorf("at least one lead minute is required")
	}
	seen := make(map[int]bool, len(r.LeadMinutes))
	for _, lead := range r.LeadMinutes {
		if lead < 1 || lead > 240 {
			return fmt.Errorf("lead minute %d must be between 1 and 240", lead)
		}
		if seen[lead] {
			return fmt.Errorf("duplicate lead minute %d", lead)
		}
		seen[lead] = true
	}
	switch r.Animation {
	case ReminderAnimationTrain, ReminderAnimationRocket, ReminderAnimationBell:
	default:
		return fmt.Errorf("unknown animation %q", r.Animation)
	}
	if r.Enabled && (r.NativeNotifications == nil || !*r.NativeNotifications) &&
		(r.PopupNotifications == nil || !*r.PopupNotifications) {
		return fmt.Errorf("enable at least one notification channel")
	}
	return nil
}

func ReminderNativeEnabled(r Reminders) bool {
	return r.NativeNotifications != nil && *r.NativeNotifications
}

func ReminderPopupEnabled(r Reminders) bool {
	return r.PopupNotifications != nil && *r.PopupNotifications
}
