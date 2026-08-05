package ui

import (
	"strings"
	"testing"
	"time"

	"pm-cli/pkg/plan"
)

func TestPlanModelShowsNormalAndConciseAlternativeClockout(t *testing.T) {
	date := time.Date(2026, time.June, 9, 0, 0, 0, 0, time.Local)
	balance := &plan.BalanceAdjustment{
		BalanceSecs:                -90 * 60,
		AppliesToday:               true,
		TargetAdjustmentSecs:       60 * 60,
		AdjustedTargetSecs:         int64((9*time.Hour + 30*time.Minute).Seconds()),
		EstimatedBalanceChangeSecs: 90 * 60,
		RemainingBalanceSecs:       0,
		Multiplier:                 1.5,
	}
	journeys := []plan.Journey{
		{
			Entry: plan.ClockSlot{Time: "08:00", Registered: true},
			Exit:  plan.ClockSlot{Time: "12:00", Registered: true},
		},
		{
			Entry: plan.ClockSlot{Time: "13:30"},
			Exit:  plan.ClockSlot{Time: "18:00"},
		},
	}
	model := NewPlanModel(
		date,
		8*time.Hour+30*time.Minute,
		nil,
		journeys,
		plan.SolvedSlot{JourneyIndex: 1, IsEntry: false},
		balance,
		"",
	)

	view := model.View()
	for _, want := range []string{"Saída 2", "18:00", "Horário alternativo", "19:00", "banco -01:30", "Meta do Dia"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
	for _, unwanted := range []string{"Saldo atual", "Ajuste hoje", "Saldo estimado", "(registrada)"} {
		if strings.Contains(view, unwanted) {
			t.Fatalf("view should not contain %q:\n%s", unwanted, view)
		}
	}
}
