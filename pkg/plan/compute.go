package plan

import (
	"fmt"
	"sort"
	"time"
)

type Record struct {
	Time time.Time
}

type Period struct {
	Enter time.Time
	Leave time.Time
}

type Suggestion struct {
	In1  time.Time
	Out1 time.Time
	In2  time.Time
	Out2 time.Time
}

func parseClock(hhmm string, date time.Time) (time.Time, error) {
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, date.Location()), nil
}

func DurationBetween(a, b time.Time) time.Duration {
	if b.Before(a) {
		return 0
	}
	return b.Sub(a)
}

func WorkedDurationFrom(records []time.Time) time.Duration {
	if len(records) < 2 {
		return 0
	}
	total := time.Duration(0)
	for i := 0; i+1 < len(records); i += 2 {
		total += DurationBetween(records[i], records[i+1])
	}
	return total
}

func minBreakFromPeriods(periods []Period) time.Duration {
	if len(periods) < 2 {
		return time.Hour
	}
	gap := periods[1].Enter.Sub(periods[0].Leave)
	if gap <= 0 {
		return time.Hour
	}
	return gap
}

func defaultOut1(in1 time.Time, periods []Period, target time.Duration) time.Time {
	if len(periods) > 0 && in1.Before(periods[0].Leave) {
		return periods[0].Leave
	}
	return in1.Add(target / 2)
}

func defaultIn2(out1 time.Time, periods []Period, minBreak time.Duration) time.Time {
	candidate := out1.Add(minBreak)
	if len(periods) > 1 && candidate.Before(periods[1].Enter) {
		return periods[1].Enter
	}
	return candidate
}

func remainingOut2(in2 time.Time, alreadyWorked time.Duration, target time.Duration, segment1 time.Duration) time.Time {
	need := target - alreadyWorked - segment1
	if need < 0 {
		need = 0
	}
	return in2.Add(need)
}

func Suggest(date time.Time, stamps []time.Time, periods []Period, target time.Duration) (Suggestion, error) {
	res := Suggestion{}
	if len(stamps) == 0 {
		return res, fmt.Errorf("no existing records")
	}

	sorted := append([]time.Time(nil), stamps...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Before(sorted[j]) })

	minBreak := minBreakFromPeriods(periods)
	already := WorkedDurationFrom(sorted)

	in1 := sorted[0]
	var out1 time.Time
	if len(sorted) >= 2 {
		out1 = sorted[1]
	}
	if out1.IsZero() {
		out1 = defaultOut1(in1, periods, target)
	}

	in2 := time.Time{}
	if len(sorted) >= 3 {
		in2 = sorted[2]
	}
	if in2.IsZero() {
		in2 = defaultIn2(out1, periods, minBreak)
	}

	seg1 := DurationBetween(in1, out1)
	out2 := time.Time{}
	if len(sorted) >= 4 {
		out2 = sorted[3]
	}
	if out2.IsZero() {
		out2 = remainingOut2(in2, already, target, seg1)
	}

	res.In1 = in1
	res.Out1 = out1
	res.In2 = in2
	res.Out2 = out2
	return res, nil
}
