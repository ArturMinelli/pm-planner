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

func Suggest(date time.Time, stamps []time.Time, periods []Period, target time.Duration, anchorsHM [4]string) (Suggestion, error) {
	res := Suggestion{}
	if len(stamps) == 0 {
		return res, fmt.Errorf("no existing records")
	}

	sorted := append([]time.Time(nil), stamps...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Before(sorted[j]) })

	minBreak := minBreakFromPeriods(periods)
	slotIx := assignStampsToPlannerSlots(sorted, date, anchorsHM)

	ts := func(slot int) time.Time {
		if ix := slotIx[slot]; ix >= 0 && ix < len(sorted) {
			return sorted[ix]
		}
		return time.Time{}
	}

	in1 := ts(0)
	out1 := ts(1)
	in2 := ts(2)
	out2 := ts(3)

	if in1.IsZero() {
		// No punch landed on the Entrada‑1 anchor: keep the planner default clock (morning anchor),
		// e.g. 12:28 goes to Saída‑1 nearest 12:00 instead of absorbing the sole punch as entrada.
		if t, err := parseClock(anchorsHM[0], date); err == nil {
			in1 = t
		} else if len(sorted) > 0 {
			in1 = sorted[0]
		}
	}
	if out1.IsZero() {
		out1 = defaultOut1(in1, periods, target)
	}
	if in2.IsZero() {
		in2 = defaultIn2(out1, periods, minBreak)
	}
	if out2.IsZero() {
		hhmm := ComputeOut2(in1.Format("15:04"), out1.Format("15:04"), in2.Format("15:04"), target, date)
		if ot, err := parseClock(hhmm, date); err == nil {
			out2 = ot
		}
	}

	res.In1 = in1
	res.Out1 = out1
	res.In2 = in2
	res.Out2 = out2
	return res, nil
}
