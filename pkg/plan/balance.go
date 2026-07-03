package plan

import (
	"math"
	"time"
)

// BalanceAdjustment describes how the current time bank changes a selected day's target.
type BalanceAdjustment struct {
	BalanceSecs                int64   `json:"balanceSecs"`
	AppliesToday               bool    `json:"appliesToday"`
	TargetAdjustmentSecs       int64   `json:"targetAdjustmentSecs"`
	AdjustedTargetSecs         int64   `json:"adjustedTargetSecs"`
	EstimatedBalanceChangeSecs int64   `json:"estimatedBalanceChangeSecs"`
	RemainingBalanceSecs       int64   `json:"remainingBalanceSecs"`
	Multiplier                 float64 `json:"multiplier"`
	Capped                     bool    `json:"capped"`
}

// CalculateBalanceAdjustment applies the time bank only when the selected date is today.
// multiplier is the credit rate for negative balance (e.g. 1.5 means 1h work = 1.5h credit).
func CalculateBalanceAdjustment(baseTarget, maxDailyExtraWork time.Duration, balanceSecs int64, selectedDate, today time.Time, multiplier float64) BalanceAdjustment {
	if baseTarget < 0 {
		baseTarget = 0
	}
	if maxDailyExtraWork < 0 {
		maxDailyExtraWork = 0
	}
	if multiplier <= 0 {
		multiplier = 1.5
	}
	out := BalanceAdjustment{
		BalanceSecs:          balanceSecs,
		AdjustedTargetSecs:   int64(baseTarget.Seconds()),
		RemainingBalanceSecs: balanceSecs,
		Multiplier:           multiplier,
	}
	if !sameCalendarDay(selectedDate, today) {
		return out
	}
	out.AppliesToday = true

	switch {
	case balanceSecs < 0:
		debt := -balanceSecs
		requiredWorkSecs := ceilToMinute(int64(math.Ceil(float64(debt) / multiplier)))
		maxExtraSecs := int64(maxDailyExtraWork.Seconds())
		if requiredWorkSecs > maxExtraSecs {
			requiredWorkSecs = maxExtraSecs
			out.Capped = true
		}
		creditSecs := int64(math.Round(float64(requiredWorkSecs) * multiplier))
		if creditSecs > debt {
			creditSecs = debt
		}
		out.TargetAdjustmentSecs = requiredWorkSecs
		out.EstimatedBalanceChangeSecs = creditSecs
		out.RemainingBalanceSecs = balanceSecs + creditSecs
	case balanceSecs > 0:
		usableSecs := floorToMinute(balanceSecs)
		targetSecs := int64(baseTarget.Seconds())
		if usableSecs > targetSecs {
			usableSecs = targetSecs
		}
		out.TargetAdjustmentSecs = -usableSecs
		out.EstimatedBalanceChangeSecs = -usableSecs
		out.RemainingBalanceSecs = balanceSecs - usableSecs
	}
	out.AdjustedTargetSecs += out.TargetAdjustmentSecs
	if out.AdjustedTargetSecs < 0 {
		out.AdjustedTargetSecs = 0
	}
	return out
}

func ceilDiv(value, divisor int64) int64 {
	if value <= 0 {
		return 0
	}
	return (value + divisor - 1) / divisor
}

func ceilToMinute(seconds int64) int64 {
	return ceilDiv(seconds, 60) * 60
}

func floorToMinute(seconds int64) int64 {
	if seconds <= 0 {
		return 0
	}
	return seconds / 60 * 60
}

func sameCalendarDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
