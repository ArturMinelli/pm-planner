package plan

import "time"

type Record struct {
	Time time.Time
}

type Period struct {
	Enter time.Time
	Leave time.Time
}

func parseClock(hhmm string, date time.Time) (time.Time, error) {
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, date.Location()), nil
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
