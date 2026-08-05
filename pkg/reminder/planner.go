package reminder

import (
	"fmt"
	"sort"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/plan"
)

// BuildDayPlan converts a workday API response into reminder slots.
func BuildDayPlan(date time.Time, wd *api.WorkDay, anchors []string, fetchedAt time.Time) (*DayPlan, error) {
	if wd == nil {
		return nil, fmt.Errorf("work day is required")
	}

	stamps := make([]time.Time, 0, len(wd.TimeCards))
	for _, card := range wd.TimeCards {
		if parsedTime, err := api.ParseHHMMOnDate(card.Time, date); err == nil {
			stamps = append(stamps, parsedTime)
		}
	}
	sort.Slice(stamps, func(i, j int) bool { return stamps[i].Before(stamps[j]) })

	target := 8*time.Hour + 30*time.Minute
	if wd.ShiftTime > 0 {
		target = time.Duration(wd.ShiftTime * float64(time.Second))
	}

	periods := make([]plan.Period, 0, len(wd.ShiftDay.Periods))
	for _, period := range wd.ShiftDay.Periods {
		enter, err1 := api.ParseHHMMOnDate(period.EnterTime, date)
		leave, err2 := api.ParseHHMMOnDate(period.LeaveTime, date)
		if err1 == nil && err2 == nil {
			periods = append(periods, plan.Period{Enter: enter, Leave: leave})
		}
	}

	var slots []Slot
	if len(stamps) == 0 {
		slots = slotsFromAnchors(date, anchors)
	} else {
		stampStrings := make([]string, len(stamps))
		for i, stamp := range stamps {
			stampStrings[i] = stamp.Format("15:04")
		}

		day, err := plan.SuggestDay(date, stampStrings, periods, target, anchors)
		if err != nil {
			return nil, err
		}

		slots = slotsFromDay(day, date)
	}

	return &DayPlan{
		Date:          date.Format("2006-01-02"),
		Target:        target,
		OriginalTimes: append([]time.Time(nil), stamps...),
		Slots:         slots,
		FetchedAt:     fetchedAt,
	}, nil
}

func slotsFromDay(day plan.Day, date time.Time) []Slot {
	slots := make([]Slot, 0, len(day.Journeys)*2)
	for journeyIndex, journey := range day.Journeys {
		entryTime, entryErr := api.ParseHHMMOnDate(journey.Entry.Time, date)
		if entryErr == nil {
			slots = append(slots, Slot{
				Key:       SlotKey(journeyIndex, false),
				Label:     SlotLabel(journeyIndex, false),
				Time:      entryTime,
				Completed: journey.Entry.Registered,
			})
		}

		exitTime, exitErr := api.ParseHHMMOnDate(journey.Exit.Time, date)
		if exitErr == nil {
			slots = append(slots, Slot{
				Key:       SlotKey(journeyIndex, true),
				Label:     SlotLabel(journeyIndex, true),
				Time:      exitTime,
				Completed: journey.Exit.Registered,
			})
		}
	}
	return slots
}

func slotsFromAnchors(date time.Time, anchors []string) []Slot {
	slots := make([]Slot, 0, len(anchors))
	for slotIndex, anchorTime := range anchors {
		journeyIndex := slotIndex / 2
		isExit := slotIndex%2 == 1

		parsedTime, err := api.ParseHHMMOnDate(anchorTime, date)
		if err != nil {
			continue
		}
		slots = append(slots, Slot{
			Key:       SlotKey(journeyIndex, isExit),
			Label:     SlotLabel(journeyIndex, isExit),
			Time:      parsedTime,
			Completed: false,
		})
	}
	return slots
}
