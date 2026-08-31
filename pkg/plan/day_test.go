package plan

import (
	"testing"
	"time"
)

func dayTestDate() time.Time {
	return time.Date(2026, time.June, 9, 0, 0, 0, 0, time.Local)
}

// --- SolveExit tests ---

func TestSolveExitAllRegisteredReturnsUnchanged(t *testing.T) {
	// Screenshot scenario: 4 registered punches → no recalc, Saída 2 stays at 14:41.
	day := Day{
		Journeys: []Journey{
			{
				Entry: ClockSlot{Time: "08:00", Registered: true},
				Exit:  ClockSlot{Time: "12:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "13:30", Registered: true},
				Exit:  ClockSlot{Time: "14:41", Registered: true},
			},
		},
		SolvedSlot: NoSolvedSlot(),
	}

	result := SolveExit(day, 8*time.Hour+30*time.Minute, -1, dayTestDate())

	if result.Journeys[1].Exit.Time != "14:41" {
		t.Errorf("Saída2 should remain 14:41, got %q", result.Journeys[1].Exit.Time)
	}
	if result.Journeys[1].Exit.Registered != true {
		t.Error("Saída2 should remain registered")
	}
}

func TestSolveExitComputesExitFromRemainingTarget(t *testing.T) {
	// Journey 1: 08:00–12:00 (4h). Target 8.5h → remaining = 4.5h → exit2 = 13:30 + 4.5h = 18:00.
	day := Day{
		Journeys: []Journey{
			{
				Entry: ClockSlot{Time: "08:00", Registered: true},
				Exit:  ClockSlot{Time: "12:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "13:30", Registered: true},
				Exit:  ClockSlot{Time: "", Registered: false},
			},
		},
		SolvedSlot: SolvedSlot{JourneyIndex: 1, IsEntry: false},
	}

	result := SolveExit(day, 8*time.Hour+30*time.Minute, 1, dayTestDate())

	if result.Journeys[1].Exit.Time != "18:00" {
		t.Errorf("expected Saída2 = 18:00, got %q", result.Journeys[1].Exit.Time)
	}
	if result.Journeys[1].Exit.Registered {
		t.Error("solved exit should not be marked registered")
	}
}

func TestSolveExitClampsToEntryWhenFixedSpanExceedsTarget(t *testing.T) {
	// Journey 1: 08:00–12:00 (4h). Target = 3h → remaining = max(3h-4h, 0) = 0 → exit = 13:00.
	day := Day{
		Journeys: []Journey{
			{
				Entry: ClockSlot{Time: "08:00", Registered: true},
				Exit:  ClockSlot{Time: "12:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "13:00", Registered: true},
				Exit:  ClockSlot{Time: "", Registered: false},
			},
		},
		SolvedSlot: SolvedSlot{JourneyIndex: 1, IsEntry: false},
	}

	result := SolveExit(day, 3*time.Hour, 1, dayTestDate())

	if result.Journeys[1].Exit.Time != "13:00" {
		t.Errorf("expected exit clamped to entry 13:00, got %q", result.Journeys[1].Exit.Time)
	}
}

func TestSolveExitPreservesOtherJourneys(t *testing.T) {
	day := Day{
		Journeys: []Journey{
			{
				Entry: ClockSlot{Time: "08:00", Registered: true},
				Exit:  ClockSlot{Time: "12:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "13:00", Registered: true},
				Exit:  ClockSlot{Time: "", Registered: false},
			},
		},
		SolvedSlot: SolvedSlot{JourneyIndex: 1, IsEntry: false},
	}

	result := SolveExit(day, 8*time.Hour, 1, dayTestDate())

	if result.Journeys[0].Entry.Time != "08:00" || result.Journeys[0].Exit.Time != "12:00" {
		t.Errorf("journey 0 was unexpectedly modified: %+v", result.Journeys[0])
	}
}

func TestSolveExitThreeJourneys(t *testing.T) {
	// Journey 1: 08:00–12:00 (4h). Journey 2: 13:00–14:00 (1h). Target 8h → remaining = 3h.
	// Journey 3 entry = 15:00, exit = 15:00 + 3h = 18:00.
	day := Day{
		Journeys: []Journey{
			{
				Entry: ClockSlot{Time: "08:00", Registered: true},
				Exit:  ClockSlot{Time: "12:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "13:00", Registered: true},
				Exit:  ClockSlot{Time: "14:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "15:00", Registered: true},
				Exit:  ClockSlot{Time: "", Registered: false},
			},
		},
		SolvedSlot: SolvedSlot{JourneyIndex: 2, IsEntry: false},
	}

	result := SolveExit(day, 8*time.Hour, 2, dayTestDate())

	if result.Journeys[2].Exit.Time != "18:00" {
		t.Errorf("expected Saída3 = 18:00, got %q", result.Journeys[2].Exit.Time)
	}
}

func TestSolveExitFourJourneys(t *testing.T) {
	day := Day{
		Journeys: []Journey{
			{
				Entry: ClockSlot{Time: "08:00", Registered: true},
				Exit:  ClockSlot{Time: "10:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "11:00", Registered: true},
				Exit:  ClockSlot{Time: "12:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "13:00", Registered: true},
				Exit:  ClockSlot{Time: "14:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "15:00", Registered: true},
				Exit:  ClockSlot{Time: "", Registered: false},
			},
		},
		SolvedSlot: SolvedSlot{JourneyIndex: 3, IsEntry: false},
	}

	result := SolveSlot(day, 8*time.Hour, SolvedSlot{JourneyIndex: 3, IsEntry: false}, dayTestDate())

	if result.Journeys[3].Exit.Time != "19:00" {
		t.Errorf("expected Saída4 = 19:00, got %q", result.Journeys[3].Exit.Time)
	}
}

// --- SolveSlot entry tests ---

func TestSolveSlotEntryTwoJourneys(t *testing.T) {
	day := Day{
		Journeys: []Journey{
			{
				Entry: ClockSlot{Time: "08:00", Registered: true},
				Exit:  ClockSlot{Time: "12:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "", Registered: false},
				Exit:  ClockSlot{Time: "18:00", Registered: false},
			},
		},
		SolvedSlot: SolvedSlot{JourneyIndex: 1, IsEntry: true},
	}

	result := SolveSlot(day, 8*time.Hour+30*time.Minute, SolvedSlot{JourneyIndex: 1, IsEntry: true}, dayTestDate())

	if result.Journeys[1].Entry.Time != "13:30" {
		t.Errorf("expected Entrada2 = 13:30, got %q", result.Journeys[1].Entry.Time)
	}
}

func TestSolveSlotEntryThreeJourneys(t *testing.T) {
	day := Day{
		Journeys: []Journey{
			{
				Entry: ClockSlot{Time: "08:00", Registered: true},
				Exit:  ClockSlot{Time: "12:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "13:00", Registered: true},
				Exit:  ClockSlot{Time: "14:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "", Registered: false},
				Exit:  ClockSlot{Time: "18:00", Registered: false},
			},
		},
		SolvedSlot: SolvedSlot{JourneyIndex: 2, IsEntry: true},
	}

	result := SolveSlot(day, 8*time.Hour, SolvedSlot{JourneyIndex: 2, IsEntry: true}, dayTestDate())

	if result.Journeys[2].Entry.Time != "15:00" {
		t.Errorf("expected Entrada3 = 15:00, got %q", result.Journeys[2].Entry.Time)
	}
}

func TestSolveSlotEntrySkipsRegisteredSlot(t *testing.T) {
	day := Day{
		Journeys: []Journey{
			{
				Entry: ClockSlot{Time: "08:00", Registered: true},
				Exit:  ClockSlot{Time: "12:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "13:30", Registered: true},
				Exit:  ClockSlot{Time: "18:00", Registered: false},
			},
		},
		SolvedSlot: SolvedSlot{JourneyIndex: 1, IsEntry: true},
	}

	result := SolveSlot(day, 8*time.Hour+30*time.Minute, SolvedSlot{JourneyIndex: 1, IsEntry: true}, dayTestDate())

	if result.Journeys[1].Entry.Time != "13:30" {
		t.Errorf("registered entry should stay 13:30, got %q", result.Journeys[1].Entry.Time)
	}
}

func TestDefaultSolvedSlotPicksLastUnregisteredExit(t *testing.T) {
	day := Day{
		Journeys: []Journey{
			{
				Entry: ClockSlot{Time: "08:00", Registered: true},
				Exit:  ClockSlot{Time: "12:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "13:30", Registered: true},
				Exit:  ClockSlot{Time: "", Registered: false},
			},
		},
	}

	slot := DefaultSolvedSlot(day)

	if slot.JourneyIndex != 1 || slot.IsEntry {
		t.Errorf("expected last unregistered exit at journey 1, got %+v", slot)
	}
}

func TestDefaultSolvedSlotPicksLastUnregisteredEntryWhenLater(t *testing.T) {
	day := Day{
		Journeys: []Journey{
			{
				Entry: ClockSlot{Time: "08:00", Registered: true},
				Exit:  ClockSlot{Time: "12:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "", Registered: false},
				Exit:  ClockSlot{Time: "18:00", Registered: true},
			},
		},
	}

	slot := DefaultSolvedSlot(day)

	if slot.JourneyIndex != 1 || !slot.IsEntry {
		t.Errorf("expected last unregistered entry at journey 1, got %+v", slot)
	}
}

func TestDefaultSolvedSlotReturnsNoneWhenAllRegistered(t *testing.T) {
	day := Day{
		Journeys: []Journey{
			{
				Entry: ClockSlot{Time: "08:00", Registered: true},
				Exit:  ClockSlot{Time: "12:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "13:30", Registered: true},
				Exit:  ClockSlot{Time: "18:00", Registered: true},
			},
		},
	}

	slot := DefaultSolvedSlot(day)

	if slot.Valid() {
		t.Errorf("expected no solved slot when all registered, got %+v", slot)
	}
}

// --- Summarize tests ---

func TestSummarizeTwoJourneys(t *testing.T) {
	date := dayTestDate()
	day := Day{
		Journeys: []Journey{
			{
				Entry: ClockSlot{Time: "08:00", Registered: true},
				Exit:  ClockSlot{Time: "12:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "13:30", Registered: true},
				Exit:  ClockSlot{Time: "18:00", Registered: false},
			},
		},
		SolvedSlot: SolvedSlot{JourneyIndex: 1, IsEntry: false},
	}
	target := 8*time.Hour + 30*time.Minute

	summary := Summarize(day, target, date)

	expectedSpan1 := int64((4 * time.Hour).Seconds())
	if summary.JourneySpanSecs[0] != expectedSpan1 {
		t.Errorf("journey 1 span: expected %d, got %d", expectedSpan1, summary.JourneySpanSecs[0])
	}

	expectedSpan2 := int64((4*time.Hour + 30*time.Minute).Seconds())
	if summary.JourneySpanSecs[1] != expectedSpan2 {
		t.Errorf("journey 2 span: expected %d, got %d", expectedSpan2, summary.JourneySpanSecs[1])
	}

	expectedTotal := int64(target.Seconds())
	if summary.TotalSpanSecs != expectedTotal {
		t.Errorf("total span: expected %d, got %d", expectedTotal, summary.TotalSpanSecs)
	}

	if summary.OvertimeSecs != 0 {
		t.Errorf("overtime: expected 0, got %d", summary.OvertimeSecs)
	}

	if len(summary.Warnings) != 0 {
		t.Errorf("expected no warnings, got %v", summary.Warnings)
	}
}

func TestSummarizeThreeJourneys(t *testing.T) {
	date := dayTestDate()
	// Journey spans: 2h + 2h + 4.5h = 8.5h = target → 0 overtime.
	day := Day{
		Journeys: []Journey{
			{
				Entry: ClockSlot{Time: "08:00", Registered: true},
				Exit:  ClockSlot{Time: "10:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "10:30", Registered: true},
				Exit:  ClockSlot{Time: "12:30", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "13:30", Registered: true},
				Exit:  ClockSlot{Time: "18:00", Registered: false},
			},
		},
		SolvedSlot: SolvedSlot{JourneyIndex: 2, IsEntry: false},
	}
	target := 8*time.Hour + 30*time.Minute

	summary := Summarize(day, target, date)

	if len(summary.JourneySpanSecs) != 3 {
		t.Fatalf("expected 3 journey spans, got %d", len(summary.JourneySpanSecs))
	}

	expectedSpans := []int64{
		int64((2 * time.Hour).Seconds()),
		int64((2 * time.Hour).Seconds()),
		int64((4*time.Hour + 30*time.Minute).Seconds()),
	}
	for journeyIndex, expectedSpan := range expectedSpans {
		if summary.JourneySpanSecs[journeyIndex] != expectedSpan {
			t.Errorf("journey %d span: expected %d, got %d", journeyIndex+1, expectedSpan, summary.JourneySpanSecs[journeyIndex])
		}
	}

	expectedTotal := int64(target.Seconds())
	if summary.TotalSpanSecs != expectedTotal {
		t.Errorf("total span: expected %d, got %d", expectedTotal, summary.TotalSpanSecs)
	}

	if summary.OvertimeSecs != 0 {
		t.Errorf("overtime: expected 0, got %d", summary.OvertimeSecs)
	}

	if len(summary.Warnings) != 0 {
		t.Errorf("expected no warnings, got %v", summary.Warnings)
	}
}

func TestSummarizeOvertimeIsNonZeroWhenTotalExceedsTarget(t *testing.T) {
	date := dayTestDate()
	day := Day{
		Journeys: []Journey{
			{
				Entry: ClockSlot{Time: "08:00", Registered: true},
				Exit:  ClockSlot{Time: "12:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "13:00", Registered: true},
				Exit:  ClockSlot{Time: "18:00", Registered: true},
			},
		},
		SolvedSlot: NoSolvedSlot(),
	}
	target := 8 * time.Hour

	summary := Summarize(day, target, date)

	// 4h + 5h = 9h; overtime = 9h - 8h = 1h
	expectedOvertime := int64((1 * time.Hour).Seconds())
	if summary.OvertimeSecs != expectedOvertime {
		t.Errorf("overtime: expected %d, got %d", expectedOvertime, summary.OvertimeSecs)
	}
}

func TestSuggestDayThreePunchesAssignsSequentiallyAndSolvesLastExit(t *testing.T) {
	date := dayTestDate()
	stamps := []string{"09:02", "13:18", "15:45"}
	anchors := []string{"08:00", "12:00", "13:30", "18:00"}
	target := 8*time.Hour + 30*time.Minute

	day, err := SuggestDay(date, stamps, nil, target, anchors)
	if err != nil {
		t.Fatalf("SuggestDay: %v", err)
	}

	if day.Journeys[0].Entry.Time != "09:02" || !day.Journeys[0].Entry.Registered {
		t.Errorf("Entrada 1: got %+v", day.Journeys[0].Entry)
	}
	if day.Journeys[0].Exit.Time != "13:18" || !day.Journeys[0].Exit.Registered {
		t.Errorf("Saída 1: got %+v", day.Journeys[0].Exit)
	}
	if day.Journeys[1].Entry.Time != "15:45" || !day.Journeys[1].Entry.Registered {
		t.Errorf("Entrada 2: got %+v", day.Journeys[1].Entry)
	}
	if day.Journeys[1].Exit.Time != "19:59" || day.Journeys[1].Exit.Registered {
		t.Errorf("Saída 2: got %+v, want unregistered 19:59", day.Journeys[1].Exit)
	}
	if day.SolvedSlot.JourneyIndex != 1 || day.SolvedSlot.IsEntry {
		t.Errorf("solved slot: got %+v, want journey 1 exit", day.SolvedSlot)
	}

	summary := Summarize(day, target, date)
	if len(summary.Warnings) != 0 {
		t.Errorf("expected no overlap warning, got %v", summary.Warnings)
	}
}

func TestSummarizeWarnsOnOrderingViolation(t *testing.T) {
	date := dayTestDate()
	day := Day{
		Journeys: []Journey{
			{
				Entry: ClockSlot{Time: "08:00", Registered: true},
				Exit:  ClockSlot{Time: "14:00", Registered: true},
			},
			{
				Entry: ClockSlot{Time: "13:30", Registered: true}, // before previous exit
				Exit:  ClockSlot{Time: "18:00", Registered: true},
			},
		},
		SolvedSlot: NoSolvedSlot(),
	}

	summary := Summarize(day, 8*time.Hour, date)

	if len(summary.Warnings) == 0 {
		t.Error("expected an ordering warning, got none")
	}
}
