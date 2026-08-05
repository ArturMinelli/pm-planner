package reminder

import (
	"fmt"
	"time"
)

// SlotKey returns the reminder key for a given journey index and whether it is an exit.
// Examples: journeyIndex=0, isExit=false → "in1"; journeyIndex=1, isExit=true → "out2"
func SlotKey(journeyIndex int, isExit bool) string {
	number := journeyIndex + 1
	if isExit {
		return fmt.Sprintf("out%d", number)
	}
	return fmt.Sprintf("in%d", number)
}

// SlotLabel returns the Portuguese label for a slot.
// Examples: journeyIndex=0, isExit=false → "Entrada 1"; journeyIndex=1, isExit=true → "Saída 2"
func SlotLabel(journeyIndex int, isExit bool) string {
	number := journeyIndex + 1
	if isExit {
		return fmt.Sprintf("Saída %d", number)
	}
	return fmt.Sprintf("Entrada %d", number)
}

// Slot is one recommended clock-in/out point for a work day.
type Slot struct {
	Key       string    `json:"key"`
	Label     string    `json:"label"`
	Time      time.Time `json:"time"`
	Completed bool      `json:"completed"`
}

// DayPlan is the reminder-ready view of the existing planner suggestion.
type DayPlan struct {
	Date          string        `json:"date"`
	Target        time.Duration `json:"target"`
	OriginalTimes []time.Time   `json:"originalTimes"`
	Slots         []Slot        `json:"slots"`
	FetchedAt     time.Time     `json:"fetchedAt"`
}

// ScheduledReminder is a concrete reminder occurrence for one slot and lead.
type ScheduledReminder struct {
	ID          string    `json:"id"`
	Date        string    `json:"date"`
	SlotKey     string    `json:"slotKey"`
	SlotLabel   string    `json:"slotLabel"`
	SlotTime    time.Time `json:"slotTime"`
	LeadMinutes int       `json:"leadMinutes"`
	FireAt      time.Time `json:"fireAt"`
}

func ReminderID(date, slotKey string, leadMinutes int) string {
	return fmt.Sprintf("%s|%s|%d", date, slotKey, leadMinutes)
}
