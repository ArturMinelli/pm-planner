package reminder

import (
	"sort"
	"time"

	"pm-cli/pkg/config"
)

const DeliveryWindow = 10 * time.Minute

// BuildSchedule returns all future and recent reminder occurrences for a day plan.
func BuildSchedule(day *DayPlan, settings config.Reminders) []ScheduledReminder {
	if day == nil || !settings.Enabled {
		return nil
	}
	out := make([]ScheduledReminder, 0, len(day.Slots)*len(settings.LeadMinutes))
	for _, slot := range day.Slots {
		if slot.Completed || slot.Time.IsZero() {
			continue
		}
		for _, lead := range settings.LeadMinutes {
			fireAt := slot.Time.Add(-time.Duration(lead) * time.Minute)
			out = append(out, ScheduledReminder{
				ID:          ReminderID(day.Date, slot.Key, lead),
				Date:        day.Date,
				SlotKey:     slot.Key,
				SlotLabel:   slot.Label,
				SlotTime:    slot.Time,
				LeadMinutes: lead,
				FireAt:      fireAt,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].FireAt.Equal(out[j].FireAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].FireAt.Before(out[j].FireAt)
	})
	return out
}

func DueReminders(day *DayPlan, settings config.Reminders, delivered DeliveryStore, suppressed map[string]bool, now time.Time) []ScheduledReminder {
	scheduled := BuildSchedule(day, settings)
	out := make([]ScheduledReminder, 0, len(scheduled))
	for _, item := range scheduled {
		if item.FireAt.After(now) {
			continue
		}
		if now.Sub(item.FireAt) > DeliveryWindow {
			continue
		}
		if delivered != nil && delivered.WasDelivered(item.ID) {
			continue
		}
		if suppressed != nil && suppressed[item.ID] {
			continue
		}
		out = append(out, item)
	}
	return out
}

func NextWake(day *DayPlan, settings config.Reminders, delivered DeliveryStore, suppressed map[string]bool, now time.Time) (time.Time, bool) {
	if due := DueReminders(day, settings, delivered, suppressed, now); len(due) > 0 {
		return now, true
	}
	for _, item := range BuildSchedule(day, settings) {
		if delivered != nil && delivered.WasDelivered(item.ID) {
			continue
		}
		if suppressed != nil && suppressed[item.ID] {
			continue
		}
		if item.FireAt.After(now) {
			return item.FireAt, true
		}
	}
	return time.Time{}, false
}
