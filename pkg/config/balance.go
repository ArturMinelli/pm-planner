package config

import (
	"fmt"
	"time"
)

const (
	DefaultMaxDailyExtraMinutes = 180
	maxDailyExtraMinutes        = 24 * 60

	DefaultBalanceCreditMultiplier = 1.5
	minBalanceCreditMultiplier     = 1.0
	maxBalanceCreditMultiplier     = 3.0
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

// ResolveBalanceCreditMultiplier returns the configured credit multiplier for the negative balance case.
// Returns DefaultBalanceCreditMultiplier when the field is unset (zero). Rejects values outside [1.0, 3.0].
func ResolveBalanceCreditMultiplier(f *File) (float64, error) {
	m := DefaultBalanceCreditMultiplier
	if f != nil && f.BalanceCreditMultiplier != 0 {
		m = f.BalanceCreditMultiplier
	}
	if m < minBalanceCreditMultiplier || m > maxBalanceCreditMultiplier {
		return 0, fmt.Errorf("must be between %.1f and %.1f", minBalanceCreditMultiplier, maxBalanceCreditMultiplier)
	}
	return m, nil
}
