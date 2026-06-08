package reminder

import (
	"fmt"
	"sort"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/plan"
)

// BuildDayPlan converts a workday API response into reminder slots.
func BuildDayPlan(date time.Time, wd *api.WorkDay, anchors [4]string, fetchedAt time.Time) (*DayPlan, error) {
	if wd == nil {
		return nil, fmt.Errorf("work day is required")
	}

	stamps := make([]time.Time, 0, len(wd.TimeCards))
	for _, c := range wd.TimeCards {
		if t, err := api.ParseHHMMOnDate(c.Time, date); err == nil {
			stamps = append(stamps, t)
		}
	}
	sort.Slice(stamps, func(i, j int) bool { return stamps[i].Before(stamps[j]) })

	target := 8*time.Hour + 30*time.Minute
	if wd.ShiftTime > 0 {
		target = time.Duration(wd.ShiftTime * float64(time.Second))
	}

	periods := make([]plan.Period, 0, len(wd.ShiftDay.Periods))
	for _, p := range wd.ShiftDay.Periods {
		enter, err1 := api.ParseHHMMOnDate(p.EnterTime, date)
		leave, err2 := api.ParseHHMMOnDate(p.LeaveTime, date)
		if err1 == nil && err2 == nil {
			periods = append(periods, plan.Period{Enter: enter, Leave: leave})
		}
	}

	var slots []Slot
	if len(stamps) == 0 {
		slots = slotsFromAnchors(date, anchors)
	} else {
		suggestion, err := plan.Suggest(date, stamps, periods, target, anchors)
		if err != nil {
			return nil, err
		}
		assigned := plan.AssignStampsToPlannerSlots(stamps, date, anchors)
		slotTimes := []time.Time{
			suggestion.In1,
			suggestion.Out1,
			suggestion.In2,
			suggestion.Out2,
		}
		slots = make([]Slot, 0, len(slotOrder))
		for i, key := range slotOrder {
			slots = append(slots, Slot{
				Key:       key,
				Label:     slotLabels[key],
				Time:      slotTimes[i],
				Completed: assigned[i] >= 0,
			})
		}
	}

	return &DayPlan{
		Date:          date.Format("2006-01-02"),
		Target:        target,
		OriginalTimes: append([]time.Time(nil), stamps...),
		Slots:         slots,
		FetchedAt:     fetchedAt,
	}, nil
}

func slotsFromAnchors(date time.Time, anchors [4]string) []Slot {
	slots := make([]Slot, 0, len(slotOrder))
	for i, key := range slotOrder {
		t, err := api.ParseHHMMOnDate(anchors[i], date)
		if err != nil {
			continue
		}
		slots = append(slots, Slot{
			Key:       key,
			Label:     slotLabels[key],
			Time:      t,
			Completed: false,
		})
	}
	return slots
}
