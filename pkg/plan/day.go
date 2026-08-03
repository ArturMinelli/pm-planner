package plan

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// ClockSlot holds a single HH:MM time value and whether it came from a real punch.
type ClockSlot struct {
	Time       string `json:"time"`       // HH:MM; empty when unknown
	Registered bool   `json:"registered"` // matched a real punch
}

// Journey groups one entry and one exit slot pair.
type Journey struct {
	Entry ClockSlot `json:"entry"`
	Exit  ClockSlot `json:"exit"`
}

// SolvedSlot identifies which clock slot is calculated from the daily target.
type SolvedSlot struct {
	JourneyIndex int  `json:"journeyIndex"`
	IsEntry      bool `json:"isEntry"`
}

// NoSolvedSlot returns the sentinel for no calculated slot.
func NoSolvedSlot() SolvedSlot {
	return SolvedSlot{JourneyIndex: -1, IsEntry: false}
}

// Valid reports whether the slot refers to a journey index.
func (slot SolvedSlot) Valid() bool {
	return slot.JourneyIndex >= 0
}

// Day holds all journeys for a single work day plus the calculated slot.
type Day struct {
	Journeys   []Journey  `json:"journeys"`
	SolvedSlot SolvedSlot `json:"solvedSlot"`
}

// Summary contains the aggregated metrics for a Day.
type Summary struct {
	Journeys        []Journey
	SolvedSlot      SolvedSlot
	JourneySpanSecs []int64
	TotalSpanSecs   int64
	OvertimeSecs    int64
	Warnings        []string
}

// addDurationToHHMM parses hhmm as "15:04", adds duration, and returns the result in "15:04" format.
func addDurationToHHMM(hhmm string, duration time.Duration) string {
	baseTime, err := time.Parse("15:04", hhmm)
	if err != nil {
		return hhmm
	}
	return baseTime.Add(duration).Format("15:04")
}

// parseBaseJourneySpan sums the work spans of all journey pairs in the base anchor slice.
func parseBaseJourneySpan(base []string) time.Duration {
	total := time.Duration(0)
	for entryIndex := 0; entryIndex+1 < len(base); entryIndex += 2 {
		entryTime, entryErr := time.Parse("15:04", base[entryIndex])
		exitTime, exitErr := time.Parse("15:04", base[entryIndex+1])
		if entryErr != nil || exitErr != nil {
			continue
		}
		span := exitTime.Sub(entryTime)
		if span > 0 {
			total += span
		}
	}
	return total
}

// ResolveAnchors returns a slice of HH:MM anchor times of length slotCount.
// Slots at indices below len(base) use the base values directly.
// Slots at indices >= 4 are derived: even (entry) = previous exit + breakDuration;
// odd (exit) = corresponding entry + remaining target share per extra journey.
func ResolveAnchors(base []string, slotCount int, breakDuration, target time.Duration) []string {
	result := make([]string, slotCount)

	baseJourneyCount := len(base) / 2
	extraJourneyCount := slotCount/2 - baseJourneyCount
	baseSpan := parseBaseJourneySpan(base)
	remaining := target - baseSpan
	if remaining < 0 {
		remaining = 0
	}

	remainingTargetShare := time.Duration(0)
	if extraJourneyCount > 0 {
		remainingTargetShare = remaining / time.Duration(extraJourneyCount)
	}

	for slotIndex := 0; slotIndex < slotCount; slotIndex++ {
		if slotIndex < len(base) {
			result[slotIndex] = base[slotIndex]
			continue
		}
		if slotIndex%2 == 0 {
			result[slotIndex] = addDurationToHHMM(result[slotIndex-1], breakDuration)
			continue
		}
		result[slotIndex] = addDurationToHHMM(result[slotIndex-1], remainingTargetShare)
	}

	return result
}

// SuggestDay builds a Day from stamps and schedule context.
// Journey count = max(2, ceil(len(stamps)/2)).
func SuggestDay(date time.Time, stamps []string, periods []Period, target time.Duration, anchors []string) (Day, error) {
	if len(stamps) == 0 {
		return Day{}, fmt.Errorf("no existing records")
	}

	sortedStamps := make([]string, len(stamps))
	copy(sortedStamps, stamps)
	sort.Strings(sortedStamps)

	journeyCount := int(math.Ceil(float64(len(sortedStamps)) / 2.0))
	if journeyCount < 2 {
		journeyCount = 2
	}
	slotCount := journeyCount * 2

	minBreak := minBreakFromPeriods(periods)
	resolvedAnchors := ResolveAnchors(anchors, slotCount, minBreak, target)
	slotAssignments := assignStampsToPlannerSlots(sortedStamps, resolvedAnchors)

	journeys := make([]Journey, journeyCount)
	for journeyIndex := 0; journeyIndex < journeyCount; journeyIndex++ {
		entrySlotIndex := journeyIndex * 2
		exitSlotIndex := journeyIndex*2 + 1

		entryAssignment := slotAssignments[entrySlotIndex]
		exitAssignment := slotAssignments[exitSlotIndex]

		entryTime := entryAssignment
		entryRegistered := entryAssignment != ""
		if !entryRegistered {
			entryTime = resolvedAnchors[entrySlotIndex]
		}

		journeys[journeyIndex] = Journey{
			Entry: ClockSlot{Time: entryTime, Registered: entryRegistered},
			Exit:  ClockSlot{Time: exitAssignment, Registered: exitAssignment != ""},
		}
	}

	day := Day{
		Journeys:   journeys,
		SolvedSlot: NoSolvedSlot(),
	}

	solvedSlot := DefaultSolvedSlot(day)
	day.SolvedSlot = solvedSlot

	return SolveSlot(day, target, solvedSlot, date), nil
}

// SolveSlot computes the time for the given slot so total worked time equals target.
// Returns the day unchanged when slot is invalid or registered.
func SolveSlot(day Day, target time.Duration, slot SolvedSlot, date time.Time) Day {
	if !slot.Valid() || slot.JourneyIndex >= len(day.Journeys) {
		return day
	}

	solvedJourney := day.Journeys[slot.JourneyIndex]
	if slot.IsEntry && solvedJourney.Entry.Registered {
		return day
	}
	if !slot.IsEntry && solvedJourney.Exit.Registered {
		return day
	}

	journeys := make([]Journey, len(day.Journeys))
	copy(journeys, day.Journeys)

	fixedSpan := time.Duration(0)
	for journeyIndex, journey := range journeys {
		if journeyIndex == slot.JourneyIndex {
			continue
		}
		entryTime, entryErr := parseClock(journey.Entry.Time, date)
		exitTime, exitErr := parseClock(journey.Exit.Time, date)
		if entryErr != nil || exitErr != nil {
			continue
		}
		span := exitTime.Sub(entryTime)
		if span > 0 {
			fixedSpan += span
		}
	}

	remaining := target - fixedSpan
	if remaining < 0 {
		remaining = 0
	}

	if slot.IsEntry {
		solvedExit, exitParseErr := parseClock(journeys[slot.JourneyIndex].Exit.Time, date)
		if exitParseErr != nil {
			return day
		}
		solvedEntryTime := solvedExit.Add(-remaining)
		if solvedEntryTime.After(solvedExit) {
			solvedEntryTime = solvedExit
		}
		journeys[slot.JourneyIndex].Entry = ClockSlot{
			Time:       solvedEntryTime.Format("15:04"),
			Registered: false,
		}
		return Day{Journeys: journeys, SolvedSlot: day.SolvedSlot}
	}

	solvedEntry, entryParseErr := parseClock(journeys[slot.JourneyIndex].Entry.Time, date)
	if entryParseErr != nil {
		return day
	}

	solvedExitTime := solvedEntry.Add(remaining)
	if solvedExitTime.Before(solvedEntry) {
		solvedExitTime = solvedEntry
	}

	journeys[slot.JourneyIndex].Exit = ClockSlot{
		Time:       solvedExitTime.Format("15:04"),
		Registered: false,
	}

	return Day{Journeys: journeys, SolvedSlot: day.SolvedSlot}
}

// SolveExit computes the exit for a journey index. Kept for tests that target exit-only solving.
func SolveExit(day Day, target time.Duration, solvedIndex int, date time.Time) Day {
	if solvedIndex < 0 {
		return day
	}
	return SolveSlot(day, target, SolvedSlot{JourneyIndex: solvedIndex, IsEntry: false}, date)
}

// DefaultSolvedSlot returns the last unregistered slot in chronological order.
func DefaultSolvedSlot(day Day) SolvedSlot {
	lastSlot := NoSolvedSlot()
	for journeyIndex, journey := range day.Journeys {
		if !journey.Entry.Registered {
			lastSlot = SolvedSlot{JourneyIndex: journeyIndex, IsEntry: true}
		}
		if !journey.Exit.Registered {
			lastSlot = SolvedSlot{JourneyIndex: journeyIndex, IsEntry: false}
		}
	}
	return lastSlot
}

// Summarize computes per-journey spans, totals, overtime, and ordering warnings for a Day.
func Summarize(day Day, target time.Duration, date time.Time) Summary {
	summary := Summary{
		Journeys:        day.Journeys,
		SolvedSlot:      day.SolvedSlot,
		JourneySpanSecs: make([]int64, len(day.Journeys)),
	}

	var warnings []string

	for journeyIndex, journey := range day.Journeys {
		entryTime, entryErr := parseClock(journey.Entry.Time, date)
		exitTime, exitErr := parseClock(journey.Exit.Time, date)

		if entryErr != nil || exitErr != nil {
			continue
		}

		if exitTime.Before(entryTime) {
			warnings = append(warnings, fmt.Sprintf(
				"journey %d: exit %s is before entry %s",
				journeyIndex+1, journey.Exit.Time, journey.Entry.Time,
			))
			continue
		}

		span := exitTime.Sub(entryTime)
		summary.JourneySpanSecs[journeyIndex] = int64(span.Seconds())
		summary.TotalSpanSecs += int64(span.Seconds())

		if journeyIndex+1 >= len(day.Journeys) {
			continue
		}
		nextJourney := day.Journeys[journeyIndex+1]
		nextEntryTime, nextEntryErr := parseClock(nextJourney.Entry.Time, date)
		if nextEntryErr != nil {
			continue
		}
		if nextEntryTime.Before(exitTime) {
			warnings = append(warnings, fmt.Sprintf(
				"journey %d entry %s is before journey %d exit %s",
				journeyIndex+2, nextJourney.Entry.Time,
				journeyIndex+1, journey.Exit.Time,
			))
		}
	}

	overtimeSecs := summary.TotalSpanSecs - int64(target.Seconds())
	if overtimeSecs < 0 {
		overtimeSecs = 0
	}
	summary.OvertimeSecs = overtimeSecs
	summary.Warnings = warnings

	return summary
}
