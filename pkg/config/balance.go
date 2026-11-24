package config

import (
	"fmt"
	"time"
)

// MaxDailyExtraWork is the fixed daily extra-work cap used by the negative-balance payoff.
const MaxDailyExtraWork = 3 * time.Hour

const (
	DefaultBalanceCreditMultiplier = 1.5
	minBalanceCreditMultiplier     = 1.0
	maxBalanceCreditMultiplier     = 3.0
)

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
