package config

import (
	"fmt"
	"time"
)

const (
	DefaultMaxDailyExtraMinutes = 180
	maxDailyExtraMinutes        = 24 * 60
)

// ResolveMaxDailyExtraWork returns the configured daily extra-work cap.
func ResolveMaxDailyExtraWork(f *File) (time.Duration, error) {
	minutes := DefaultMaxDailyExtraMinutes
	if f != nil && f.MaxDailyExtraMinutes != 0 {
		minutes = f.MaxDailyExtraMinutes
	}
	if minutes < 1 || minutes > maxDailyExtraMinutes {
		return 0, fmt.Errorf("must be between 1 and %d minutes", maxDailyExtraMinutes)
	}
	return time.Duration(minutes) * time.Minute, nil
}
