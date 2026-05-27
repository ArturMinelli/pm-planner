package plan

import "time"

// ComputeOut2 returns the computed second exit clock (HH:mm) for a target work span,
// given first segment boundaries and second entry. Mirrors historical CLI behavior.
func ComputeOut2(in1, out1, in2 string, target time.Duration, date time.Time) string {
	parse := func(s string) time.Time {
		t, _ := time.ParseInLocation("15:04", s, date.Location())
		return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, date.Location())
	}
	in1T, out1T, in2T := parse(in1), parse(out1), parse(in2)
	first := out1T.Sub(in1T)
	need := target - first
	if need < 0 {
		need = 0
	}
	if in2T.IsZero() {
		return "--:--"
	}
	return in2T.Add(need).Format("15:04")
}

// SegmentDurations sums first segment (in1→out1), second (in2→out2), total, and overtime over target (non-negative parts).
func SegmentDurations(in1, out1, in2, out2 string, target time.Duration, date time.Time) (first, second, total, overtime time.Duration) {
	parse := func(s string) time.Time {
		t, err := time.ParseInLocation("15:04", s, date.Location())
		if err != nil {
			return time.Time{}
		}
		return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, date.Location())
	}
	a, b, c := parse(in1), parse(out1), parse(in2)
	d := parse(out2)

	first = durationBetweenClocks(a, b)
	second = durationBetweenClocks(c, d)
	total = first + second
	var extra time.Duration
	if total > target {
		extra = total - target
	}
	return first, second, total, extra
}

func durationBetweenClocks(start, end time.Time) time.Duration {
	if start.IsZero() || end.IsZero() || end.Before(start) {
		return 0
	}
	return end.Sub(start)
}
