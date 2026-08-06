package app

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/config"
	"pm-cli/pkg/plan"
)

const defaultPlannerTarget = 8*time.Hour + 30*time.Minute

// PlannerPayload is JSON-friendly state after loading work day data (desktop + shared logic).
type PlannerPayload struct {
	Date           string                  `json:"date"`
	BaseTargetSecs int64                   `json:"baseTargetSecs"`
	Balance        *plan.BalanceAdjustment `json:"balance,omitempty"`
	BalanceError   string                  `json:"balanceError,omitempty"`
	OriginalTimes  []string                `json:"originalTimes"`
	Journeys       []plan.Journey          `json:"journeys"`
	SolvedSlot     plan.SolvedSlot         `json:"solvedSlot"`
	OriginalsLine  string                  `json:"originalsLine"`
	LoadWarning    string                  `json:"loadWarning,omitempty"`
}

// PlannerSummary is returned when recalculating from editable journey inputs.
type PlannerSummary struct {
	Journeys        []plan.Journey `json:"journeys"`
	SolvedSlot      plan.SolvedSlot `json:"solvedSlot"`
	JourneySpanSecs []int64        `json:"journeySpanSecs"`
	TotalSpanSecs   int64          `json:"totalSpanSecs"`
	OvertimeSecs    int64          `json:"overtimeSecs"`
	AlternativeTime string         `json:"alternativeTime,omitempty"`
	Warnings        []string       `json:"warnings,omitempty"`
}

// FetchPlannerPayload loads the work day and builds suggestion fields for the desktop planner.
func FetchPlannerPayload(ctx context.Context, client *api.Client, dateStr string) (*PlannerPayload, error) {
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	wd, err := client.FetchWorkDay(ctx, dateStr)
	if err != nil {
		return nil, err
	}

	loc := time.Now().Location()
	date, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return nil, fmt.Errorf("date: %w", err)
	}

	stamps := make([]time.Time, 0, len(wd.TimeCards))
	for _, c := range wd.TimeCards {
		if t, err := api.ParseHHMMOnDate(c.Time, date); err == nil {
			stamps = append(stamps, t)
		}
	}
	sort.Slice(stamps, func(i, j int) bool { return stamps[i].Before(stamps[j]) })

	periods := make([]plan.Period, 0, len(wd.ShiftDay.Periods))
	for _, p := range wd.ShiftDay.Periods {
		enter, err1 := api.ParseHHMMOnDate(p.EnterTime, date)
		leave, err2 := api.ParseHHMMOnDate(p.LeaveTime, date)
		if err1 == nil && err2 == nil {
			periods = append(periods, plan.Period{Enter: enter, Leave: leave})
		}
	}

	baseTarget := defaultPlannerTarget
	if wd.ShiftTime > 0 {
		baseTarget = time.Duration(wd.ShiftTime * float64(time.Second))
	}
	cfg, _ := config.Read("")
	maxDailyExtraWork, err := config.ResolveMaxDailyExtraWork(cfg)
	if err != nil {
		maxDailyExtraWork = time.Duration(config.DefaultMaxDailyExtraMinutes) * time.Minute
	}
	balanceCreditMultiplier, err := config.ResolveBalanceCreditMultiplier(cfg)
	if err != nil {
		balanceCreditMultiplier = config.DefaultBalanceCreditMultiplier
	}
	var balanceAdjustment *plan.BalanceAdjustment
	var balanceErr string
	if balance, err := client.FetchEmployeeBalance(ctx); err != nil {
		balanceErr = err.Error()
	} else {
		adjustment := plan.CalculateBalanceAdjustment(baseTarget, maxDailyExtraWork, balance.TimeBalanceSecs, date, time.Now().In(loc), balanceCreditMultiplier)
		balanceAdjustment = &adjustment
	}

	anchors, _ := config.ResolvePlannerAnchors(cfg)

	stampStrings := make([]string, len(stamps))
	for i, t := range stamps {
		stampStrings[i] = t.Format("15:04")
	}

	day, err := plan.SuggestDay(date, stampStrings, periods, baseTarget, anchors)
	if err != nil {
		return nil, err
	}

	return &PlannerPayload{
		Date:           dateStr,
		BaseTargetSecs: int64(baseTarget.Seconds()),
		Balance:        balanceAdjustment,
		BalanceError:   balanceErr,
		OriginalTimes:  OriginalStampStrings(stamps),
		Journeys:       day.Journeys,
		SolvedSlot:     day.SolvedSlot,
		OriginalsLine:  FormatOriginalStamps(stamps),
	}, nil
}

// BuildDefaultsPlannerPayload builds a settings-based planner payload when the API is unavailable.
// Two unregistered journeys use resolved anchor times; solved exit is the last journey; no balance.
func BuildDefaultsPlannerPayload(dateStr string) (*PlannerPayload, error) {
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	loc := time.Now().Location()
	date, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return nil, fmt.Errorf("date: %w", err)
	}

	cfg, _ := config.Read("")
	anchors, _ := config.ResolvePlannerAnchors(cfg)
	if len(anchors) < 4 {
		builtin := plan.BuiltinAnchors()
		anchors = []string{builtin[0], builtin[1], builtin[2], builtin[3]}
	}

	journeys := []plan.Journey{
		{
			Entry: plan.ClockSlot{Time: anchors[0], Registered: false},
			Exit:  plan.ClockSlot{Time: anchors[1], Registered: false},
		},
		{
			Entry: plan.ClockSlot{Time: anchors[2], Registered: false},
			Exit:  plan.ClockSlot{Time: anchors[3], Registered: false},
		},
	}
	solvedSlot := plan.SolvedSlot{JourneyIndex: 1, IsEntry: false}
	day := plan.SolveSlot(
		plan.Day{Journeys: journeys, SolvedSlot: solvedSlot},
		defaultPlannerTarget,
		solvedSlot,
		date,
	)

	return &PlannerPayload{
		Date:           dateStr,
		BaseTargetSecs: int64(defaultPlannerTarget.Seconds()),
		OriginalTimes:  []string{},
		Journeys:       day.Journeys,
		SolvedSlot:     solvedSlot,
		OriginalsLine:  "(nenhum)",
	}, nil
}

// RecalculatePlanner recomputes the solved slot and segment summaries from journey inputs.
func RecalculatePlanner(date time.Time, baseTargetSecs int64, balance *plan.BalanceAdjustment, journeys []plan.Journey, solvedSlot plan.SolvedSlot) (*PlannerSummary, error) {
	target := time.Duration(baseTargetSecs) * time.Second
	if target <= 0 {
		target = defaultPlannerTarget
	}

	day := plan.Day{Journeys: journeys, SolvedSlot: solvedSlot}
	day = plan.SolveSlot(day, target, solvedSlot, date)

	summary := plan.Summarize(day, target, date)

	alternativeTime := computeAlternativeTime(day, baseTargetSecs, solvedSlot, balance, date)

	return &PlannerSummary{
		Journeys:        summary.Journeys,
		SolvedSlot:      summary.SolvedSlot,
		JourneySpanSecs: summary.JourneySpanSecs,
		TotalSpanSecs:   summary.TotalSpanSecs,
		OvertimeSecs:    summary.OvertimeSecs,
		AlternativeTime: alternativeTime,
		Warnings:        summary.Warnings,
	}, nil
}

func computeAlternativeTime(day plan.Day, baseTargetSecs int64, solvedSlot plan.SolvedSlot, balance *plan.BalanceAdjustment, date time.Time) string {
	if balance == nil || !balance.AppliesToday || balance.TargetAdjustmentSecs == 0 {
		return ""
	}
	if !solvedSlot.Valid() {
		return ""
	}
	adjustedTargetSecs := baseTargetSecs + balance.TargetAdjustmentSecs
	adjustedTarget := time.Duration(adjustedTargetSecs) * time.Second
	altDay := plan.SolveSlot(day, adjustedTarget, solvedSlot, date)
	if solvedSlot.JourneyIndex >= len(altDay.Journeys) {
		return ""
	}
	solvedJourney := altDay.Journeys[solvedSlot.JourneyIndex]
	if solvedSlot.IsEntry {
		return solvedJourney.Entry.Time
	}
	return solvedJourney.Exit.Time
}

// FormatOriginalStamps formats clock stamps as one journey per line (entry — exit).
func FormatOriginalStamps(stamps []time.Time) string {
	return FormatOriginalStampStrings(OriginalStampStrings(stamps))
}

// FormatOriginalStampStrings formats HH:mm stamps as one journey per line (entry — exit).
func FormatOriginalStampStrings(stamps []string) string {
	if len(stamps) == 0 {
		return "(nenhum)"
	}

	lines := make([]string, 0, (len(stamps)+1)/2)
	for index := 0; index < len(stamps); index += 2 {
		if index+1 < len(stamps) {
			lines = append(lines, stamps[index]+" — "+stamps[index+1])
			continue
		}
		lines = append(lines, stamps[index])
	}
	return strings.Join(lines, "\n")
}

// OriginalStampStrings is the same times as HH:mm list.
func OriginalStampStrings(stamps []time.Time) []string {
	out := make([]string, 0, len(stamps))
	for _, t := range stamps {
		out = append(out, t.Format("15:04"))
	}
	return out
}
