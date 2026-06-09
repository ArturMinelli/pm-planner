package cmd

import (
	"strings"
	"testing"
	"time"

	"pm-cli/pkg/plan"
)

func TestFormatAlternativeClockout(t *testing.T) {
	balance := &plan.BalanceAdjustment{
		BalanceSecs:          -5400,
		AppliesToday:         true,
		TargetAdjustmentSecs: 3600,
	}
	got := formatAlternativeClockout("19:00", balance)
	if got != "Saída 2 alternativa: 19:00 (banco -01:30)" {
		t.Fatalf("alternative line: %q", got)
	}

	balance.AppliesToday = false
	if got := formatAlternativeClockout("19:00", balance); got != "" {
		t.Fatalf("non-today alternative should be hidden: %q", got)
	}
}

func TestBalanceUnavailableTodayIsConcise(t *testing.T) {
	today := time.Now()
	got := balanceUnavailableToday(today, "upstream details")
	if got == "" || strings.Contains(got, "upstream details") {
		t.Fatalf("warning should be concise: %q", got)
	}
	if got := balanceUnavailableToday(today.AddDate(0, 0, -1), "upstream details"); got != "" {
		t.Fatalf("non-today warning should be hidden: %q", got)
	}
}
