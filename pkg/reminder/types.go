package reminder

import (
	"fmt"
	"time"
)

const (
	SlotIn1  = "in1"
	SlotOut1 = "out1"
	SlotIn2  = "in2"
	SlotOut2 = "out2"
)

var slotLabels = map[string]string{
	SlotIn1:  "Entrada 1",
	SlotOut1: "Saída 1",
	SlotIn2:  "Entrada 2",
	SlotOut2: "Saída 2",
}

var slotOrder = []string{SlotIn1, SlotOut1, SlotIn2, SlotOut2}

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
