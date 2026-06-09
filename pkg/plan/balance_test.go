package plan

import (
	"testing"
	"time"
)

func TestCalculateBalanceAdjustmentNegativeBalanceUsesMultiplier(t *testing.T) {
	today := balanceTestDate()
	got := CalculateBalanceAdjustment(8*time.Hour+30*time.Minute, 3*time.Hour, -90*60, today, today)
	if got.TargetAdjustmentSecs != 60*60 {
		t.Fatalf("target adjustment: got %d", got.TargetAdjustmentSecs)
	}
	if got.EstimatedBalanceChangeSecs != 90*60 || got.RemainingBalanceSecs != 0 {
		t.Fatalf("balance result: %#v", got)
	}
}

func TestCalculateBalanceAdjustmentCapsActualExtraWork(t *testing.T) {
	today := balanceTestDate()
	got := CalculateBalanceAdjustment(8*time.Hour, 2*time.Hour, -8*60*60, today, today)
	if got.TargetAdjustmentSecs != int64((2*time.Hour).Seconds()) || !got.Capped {
		t.Fatalf("cap result: %#v", got)
	}
	if got.EstimatedBalanceChangeSecs != int64((3 * time.Hour).Seconds()) {
		t.Fatalf("estimated credit: got %d", got.EstimatedBalanceChangeSecs)
	}
}

func TestCalculateBalanceAdjustmentPositiveBalanceUsesOneToOne(t *testing.T) {
	today := balanceTestDate()
	got := CalculateBalanceAdjustment(8*time.Hour, 3*time.Hour, 90*60, today, today)
	if got.TargetAdjustmentSecs != -90*60 || got.AdjustedTargetSecs != int64((6*time.Hour+30*time.Minute).Seconds()) {
		t.Fatalf("positive balance result: %#v", got)
	}
}

func TestCalculateBalanceAdjustmentRoundsByDirectionAndFloorsTarget(t *testing.T) {
	today := balanceTestDate()
	debt := CalculateBalanceAdjustment(8*time.Hour, 3*time.Hour, -91, today, today)
	if debt.TargetAdjustmentSecs != 120 {
		t.Fatalf("debt should round actual work up to minute: %#v", debt)
	}
	credit := CalculateBalanceAdjustment(30*time.Second, 3*time.Hour, 119, today, today)
	if credit.TargetAdjustmentSecs != -30 || credit.AdjustedTargetSecs != 0 {
		t.Fatalf("credit should floor to target zero: %#v", credit)
	}
}

func TestCalculateBalanceAdjustmentDoesNotApplyOutsideToday(t *testing.T) {
	today := balanceTestDate()
	got := CalculateBalanceAdjustment(8*time.Hour, 3*time.Hour, -90*60, today.AddDate(0, 0, 1), today)
	if got.AppliesToday || got.TargetAdjustmentSecs != 0 || got.RemainingBalanceSecs != -90*60 {
		t.Fatalf("non-today result: %#v", got)
	}
}

func TestCalculateBalanceAdjustmentZeroBalance(t *testing.T) {
	today := balanceTestDate()
	got := CalculateBalanceAdjustment(8*time.Hour, 3*time.Hour, 0, today, today)
	if !got.AppliesToday || got.TargetAdjustmentSecs != 0 || got.AdjustedTargetSecs != int64((8*time.Hour).Seconds()) {
		t.Fatalf("zero balance result: %#v", got)
	}
}

func balanceTestDate() time.Time {
	return time.Date(2026, time.June, 9, 12, 0, 0, 0, time.Local)
}
