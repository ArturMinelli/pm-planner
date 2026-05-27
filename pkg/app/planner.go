package app

import (
	"context"
	"fmt"
	"sort"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/plan"
)

const defaultPlannerTarget = 8*time.Hour + 30*time.Minute

// PlannerPayload is JSON-friendly state after loading work day data (desktop + shared logic).
type PlannerPayload struct {
	Date             string   `json:"date"`
	TargetSecs       int64    `json:"targetSecs"`
	OriginalTimes    []string `json:"originalTimes"`
	In1              string   `json:"in1"`
	Out1             string   `json:"out1"`
	In2              string   `json:"in2"`
	Out2             string   `json:"out2"`
	OriginalsLine    string   `json:"originalsLine"`
}

// PlannerSummary is returned when recalculating from editable clock strings.
type PlannerSummary struct {
	Out2            string `json:"out2"`
	FirstSpanSecs   int64  `json:"firstSpanSecs"`
	SecondSpanSecs  int64  `json:"secondSpanSecs"`
	TotalSpanSecs   int64  `json:"totalSpanSecs"`
	OvertimeSecs    int64  `json:"overtimeSecs"`
}

// FetchPlannerPayload loads the work day and builds suggestion fields (same pipeline as CLI `pm plan`).
func FetchPlannerPayload(ctx context.Context, client *api.Client, dateStr string) (*PlannerPayload, error) {
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	wd, err := client.FetchWorkDay(ctx, dateStr)
	if err != nil {
		return nil, err
	}

	loc := time.Now().Location()
	date, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return nil, fmt.Errorf("date: %w", err)
	}

	stamps := make([]time.Time, 0, len(wd.TimeCards))
	for _, c := range wd.TimeCards {
		if t, err := api.ParseHHMMOnDate(c.Time, date); err == nil {
			stamps = append(stamps, t)
		}
	}
	sort.Slice(stamps, func(i, j int) bool { return stamps[i].Before(stamps[j]) })

	periods := make([]plan.Period, 0, len(wd.ShiftDay.Periods))
	for _, p := range wd.ShiftDay.Periods {
		enter, err1 := api.ParseHHMMOnDate(p.EnterTime, date)
		leave, err2 := api.ParseHHMMOnDate(p.LeaveTime, date)
		if err1 == nil && err2 == nil {
			periods = append(periods, plan.Period{Enter: enter, Leave: leave})
		}
	}

	target := defaultPlannerTarget
	if wd.ShiftTime > 0 {
		target = time.Duration(wd.ShiftTime * float64(time.Second))
	}

	sug, err := plan.Suggest(date, stamps, periods, target)
	if err != nil {
		return nil, err
	}

	in1Str := sug.In1.Format("15:04")
	out1Str := sug.Out1.Format("15:04")
	in2Str := sug.In2.Format("15:04")
	out2Str := sug.Out2.Format("15:04")
	if sug.Out2.IsZero() {
		out2Str = plan.ComputeOut2(in1Str, out1Str, in2Str, target, date)
	}

	orig := FormatOriginalStamps(stamps)

	return &PlannerPayload{
		Date:          dateStr,
		TargetSecs:    int64(target.Seconds()),
		OriginalTimes: OriginalStampStrings(stamps),
		In1:           in1Str,
		Out1:          out1Str,
		In2:           in2Str,
		Out2:          out2Str,
		OriginalsLine: orig,
	}, nil
}

// RecalculatePlanner updates out2 and segment summaries from edited HH:mm inputs.
func RecalculatePlanner(dateStr string, targetSecs int64, in1, out1, in2 string) (*PlannerSummary, error) {
	loc := time.Now().Location()
	date, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return nil, fmt.Errorf("date: %w", err)
	}
	target := time.Duration(targetSecs) * time.Second
	if target <= 0 {
		target = defaultPlannerTarget
	}
	out2 := plan.ComputeOut2(in1, out1, in2, target, date)

	first, second, total, extra := plan.SegmentDurations(in1, out1, in2, out2, target, date)
	return &PlannerSummary{
		Out2:           out2,
		FirstSpanSecs:   int64(first.Seconds()),
		SecondSpanSecs:  int64(second.Seconds()),
		TotalSpanSecs:   int64(total.Seconds()),
		OvertimeSecs:    int64(extra.Seconds()),
	}, nil
}

// FormatOriginalStamps matches CLI wording for huh note / display.
func FormatOriginalStamps(stamps []time.Time) string {
	if len(stamps) == 0 {
		return "(nenhum)"
	}
	s := ""
	for i, t := range stamps {
		if i > 0 {
			s += ", "
		}
		s += t.Format("15:04")
	}
	return s
}

// OriginalStampStrings is the same times as HH:mm list.
func OriginalStampStrings(stamps []time.Time) []string {
	out := make([]string, 0, len(stamps))
	for _, t := range stamps {
		out = append(out, t.Format("15:04"))
	}
	return out
}
