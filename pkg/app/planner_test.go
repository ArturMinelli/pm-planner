package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/config"
	"pm-cli/pkg/plan"
)

func TestFetchPlannerPayloadAppliesTodayBalance(t *testing.T) {
	setupPlannerSession(t)
	today := time.Now().Format("2006-01-02")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/time_card_control/current/work_days/" + today:
			_, _ = w.Write([]byte(`{"work_day":{"shift_time":30600,"shift_day":{"periods":[]},"time_cards":[{"time":"08:00"}]}}`))
		case "/employees/my_time_break":
			_, _ = w.Write([]byte(`{"employee":{"id":42}}`))
		case "/employees/statuses/42":
			_, _ = w.Write([]byte(`{"statuses":{"time_balance":-5400}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	payload, err := FetchPlannerPayload(context.Background(), plannerTestClient(srv.URL), today)
	if err != nil {
		t.Fatal(err)
	}
	if payload.BaseTargetSecs != 30600 {
		t.Fatalf("base target: %d", payload.BaseTargetSecs)
	}
	if payload.Balance == nil || payload.Balance.TargetAdjustmentSecs != 3600 {
		t.Fatalf("balance: %#v", payload.Balance)
	}
	if !payload.SolvedSlot.Valid() || payload.SolvedSlot.JourneyIndex >= len(payload.Journeys) || payload.Journeys[payload.SolvedSlot.JourneyIndex].Exit.Time == "" {
		t.Fatal("normal clockout should remain available")
	}
}

func TestFetchPlannerPayloadKeepsPlanningWhenBalanceFails(t *testing.T) {
	setupPlannerSession(t)
	today := time.Now().Format("2006-01-02")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/time_card_control/current/work_days/" + today:
			_, _ = w.Write([]byte(`{"work_day":{"shift_time":30600,"shift_day":{"periods":[]},"time_cards":[{"time":"08:00"}]}}`))
		case "/employees/my_time_break":
			http.Error(w, "balance unavailable", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	payload, err := FetchPlannerPayload(context.Background(), plannerTestClient(srv.URL), today)
	if err != nil {
		t.Fatal(err)
	}
	if payload.Balance != nil {
		t.Fatalf("expected no balance: %#v", payload.Balance)
	}
	if payload.BalanceError == nil {
		t.Fatal("expected balance error")
	}
	if !payload.SolvedSlot.Valid() || payload.SolvedSlot.JourneyIndex >= len(payload.Journeys) || payload.Journeys[payload.SolvedSlot.JourneyIndex].Exit.Time == "" {
		t.Fatal("normal clockout should remain available")
	}
}

func TestFetchPlannerPayloadUsesConfiguredDailyExtraCap(t *testing.T) {
	setupPlannerSession(t)
	if err := config.Save("", &config.File{MaxDailyExtraMinutes: 90}); err != nil {
		t.Fatal(err)
	}
	today := time.Now().Format("2006-01-02")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/time_card_control/current/work_days/" + today:
			_, _ = w.Write([]byte(`{"work_day":{"shift_time":30600,"shift_day":{"periods":[]},"time_cards":[{"time":"08:00"}]}}`))
		case "/employees/my_time_break":
			_, _ = w.Write([]byte(`{"employee":{"id":42}}`))
		case "/employees/statuses/42":
			_, _ = w.Write([]byte(`{"statuses":{"time_balance":-28800}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	payload, err := FetchPlannerPayload(context.Background(), plannerTestClient(srv.URL), today)
	if err != nil {
		t.Fatal(err)
	}
	if payload.Balance == nil || payload.Balance.TargetAdjustmentSecs != 90*60 || !payload.Balance.Capped {
		t.Fatalf("configured cap not applied: %#v", payload.Balance)
	}
}

func TestRecalculatePlannerKeepsNormalClockoutPrimary(t *testing.T) {
	balance := plan.BalanceAdjustment{
		AppliesToday:         true,
		TargetAdjustmentSecs: 3600,
		AdjustedTargetSecs:   9.5 * 60 * 60,
	}
	date, _ := time.Parse("2006-01-02", "2026-06-09")
	journeys := []plan.Journey{
		{Entry: plan.ClockSlot{Time: "08:00", Registered: true}, Exit: plan.ClockSlot{Time: "12:00", Registered: true}},
		{Entry: plan.ClockSlot{Time: "13:30", Registered: true}, Exit: plan.ClockSlot{Registered: false}},
	}
	solvedSlot := plan.SolvedSlot{JourneyIndex: 1, IsEntry: false}
	got, err := RecalculatePlanner(date, 8.5*60*60, &balance, journeys, solvedSlot)
	if err != nil {
		t.Fatal(err)
	}
	if got.SolvedSlot != solvedSlot || got.Journeys[1].Exit.Time != "18:00" {
		t.Fatalf("normal clockout: %#v", got)
	}
	if got.AlternativeTime != "19:00" {
		t.Fatalf("alternative time: got %q, want 19:00", got.AlternativeTime)
	}
	if got.TotalSpanSecs != 8.5*60*60 || got.OvertimeSecs != 0 {
		t.Fatalf("normal totals changed by balance: %#v", got)
	}
}

func TestRecalculatePlannerSupportsEarlierAndCappedAlternatives(t *testing.T) {
	tests := []struct {
		name        string
		balance     plan.BalanceAdjustment
		alternative string
	}{
		{
			name: "positive credit",
			balance: plan.BalanceAdjustment{
				AppliesToday:         true,
				TargetAdjustmentSecs: -60 * 60,
				AdjustedTargetSecs:   7.5 * 60 * 60,
			},
			alternative: "17:00",
		},
		{
			name: "three hour cap",
			balance: plan.BalanceAdjustment{
				AppliesToday:         true,
				TargetAdjustmentSecs: 3 * 60 * 60,
				AdjustedTargetSecs:   11.5 * 60 * 60,
				Capped:               true,
			},
			alternative: "21:00",
		},
	}
	date, _ := time.Parse("2006-01-02", "2026-06-09")
	journeys := []plan.Journey{
		{Entry: plan.ClockSlot{Time: "08:00", Registered: true}, Exit: plan.ClockSlot{Time: "12:00", Registered: true}},
		{Entry: plan.ClockSlot{Time: "13:30", Registered: true}, Exit: plan.ClockSlot{Registered: false}},
	}
	solvedSlot := plan.SolvedSlot{JourneyIndex: 1, IsEntry: false}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RecalculatePlanner(date, 8.5*60*60, &tt.balance, journeys, solvedSlot)
			if err != nil {
				t.Fatal(err)
			}
			if got.Journeys[1].Exit.Time != "18:00" {
				t.Fatalf("normal clockout: %q, want 18:00", got.Journeys[1].Exit.Time)
			}
			if got.AlternativeTime != tt.alternative {
				t.Fatalf("alternative time: got %q, want %q", got.AlternativeTime, tt.alternative)
			}
			if got.TotalSpanSecs != 8.5*60*60 || got.OvertimeSecs != 0 {
				t.Fatalf("normal totals changed by alternative: %#v", got)
			}
		})
	}
}

func TestRecalculatePlannerHidesNonTodayAlternative(t *testing.T) {
	balance := plan.BalanceAdjustment{
		AppliesToday:         false,
		TargetAdjustmentSecs: -3600,
		AdjustedTargetSecs:   7.5 * 60 * 60,
	}
	date, _ := time.Parse("2006-01-02", "2026-06-08")
	journeys := []plan.Journey{
		{Entry: plan.ClockSlot{Time: "08:00", Registered: true}, Exit: plan.ClockSlot{Time: "12:00", Registered: true}},
		{Entry: plan.ClockSlot{Time: "13:30", Registered: true}, Exit: plan.ClockSlot{Registered: false}},
	}
	solvedSlot := plan.SolvedSlot{JourneyIndex: 1, IsEntry: false}
	got, err := RecalculatePlanner(date, 8.5*60*60, &balance, journeys, solvedSlot)
	if err != nil {
		t.Fatal(err)
	}
	if got.Journeys[1].Exit.Time != "18:00" {
		t.Fatalf("normal clockout: %q, want 18:00", got.Journeys[1].Exit.Time)
	}
	if got.AlternativeTime != "" {
		t.Fatalf("non-today should have no alternative: %q", got.AlternativeTime)
	}
}

func TestBuildDefaultsPlannerPayloadShape(t *testing.T) {
	setupPlannerSession(t)

	payload, err := BuildDefaultsPlannerPayload("2026-06-09")
	if err != nil {
		t.Fatal(err)
	}
	if payload.Date != "2026-06-09" {
		t.Fatalf("date: %q", payload.Date)
	}
	if payload.BaseTargetSecs != int64(defaultPlannerTarget.Seconds()) {
		t.Fatalf("base target: %d", payload.BaseTargetSecs)
	}
	if payload.Balance != nil || payload.BalanceError != nil {
		t.Fatalf("defaults should omit balance: %#v / %#v", payload.Balance, payload.BalanceError)
	}
	if payload.OriginalsLine != "" {
		t.Fatalf("originals: %q", payload.OriginalsLine)
	}
	if len(payload.OriginalTimes) != 0 {
		t.Fatalf("original times: %#v", payload.OriginalTimes)
	}
	if len(payload.Journeys) != 2 {
		t.Fatalf("journey count: %d", len(payload.Journeys))
	}
	for journeyIndex, journey := range payload.Journeys {
		if journey.Entry.Registered || journey.Exit.Registered {
			t.Fatalf("journey %d should be unregistered: %#v", journeyIndex, journey)
		}
		if journey.Entry.Time == "" || journey.Exit.Time == "" {
			t.Fatalf("journey %d missing times: %#v", journeyIndex, journey)
		}
	}
	if !payload.SolvedSlot.Valid() || payload.SolvedSlot.JourneyIndex != 1 || payload.SolvedSlot.IsEntry {
		t.Fatalf("solved slot: %#v", payload.SolvedSlot)
	}
	if payload.LoadWarning != nil {
		t.Fatalf("builder should not set load warning: %#v", payload.LoadWarning)
	}
}

func TestFormatOriginalStampStrings(t *testing.T) {
	tests := []struct {
		name   string
		stamps []string
		want   string
	}{
		{
			name:   "empty",
			stamps: nil,
			want:   "",
		},
		{
			name:   "single stamp",
			stamps: []string{"08:00"},
			want:   "08:00",
		},
		{
			name:   "one journey",
			stamps: []string{"08:00", "18:00"},
			want:   "08:00 — 18:00",
		},
		{
			name:   "two journeys",
			stamps: []string{"08:00", "12:00", "13:00", "18:00"},
			want:   "08:00 — 12:00\n13:00 — 18:00",
		},
		{
			name:   "odd stamp count",
			stamps: []string{"08:00", "12:00", "13:00"},
			want:   "08:00 — 12:00\n13:00",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if got := FormatOriginalStampStrings(testCase.stamps); got != testCase.want {
				t.Fatalf("FormatOriginalStampStrings() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestBuildDefaultsPlannerPayloadUsesConfiguredAnchors(t *testing.T) {
	setupPlannerSession(t)
	if err := config.Save("", &config.File{
		Planner: &config.PlannerAnchors{
			In1:  "07:30",
			Out1: "11:30",
			In2:  "12:30",
			Out2: "17:00",
		},
	}); err != nil {
		t.Fatal(err)
	}

	payload, err := BuildDefaultsPlannerPayload("2026-06-09")
	if err != nil {
		t.Fatal(err)
	}
	if payload.Journeys[0].Entry.Time != "07:30" || payload.Journeys[0].Exit.Time != "11:30" {
		t.Fatalf("journey 1: %#v", payload.Journeys[0])
	}
	if payload.Journeys[1].Entry.Time != "12:30" {
		t.Fatalf("journey 2 entry: %#v", payload.Journeys[1])
	}
	if !payload.SolvedSlot.Valid() || payload.SolvedSlot.JourneyIndex != 1 || payload.SolvedSlot.IsEntry {
		t.Fatalf("solved slot: %#v", payload.SolvedSlot)
	}
	// Solved exit is recomputed from target; with these anchors remaining is 4h30 → 17:00.
	if payload.Journeys[1].Exit.Time != "17:00" {
		t.Fatalf("solved exit time: %q", payload.Journeys[1].Exit.Time)
	}
}

func plannerTestClient(baseURL string) *api.Client {
	return &api.Client{
		HTTP:    http.DefaultClient,
		BaseURL: baseURL,
		AuthHeaders: func() (map[string]string, error) {
			return map[string]string{"Access-Token": "test"}, nil
		},
	}
}

func setupPlannerSession(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, ".config"))
	configDir, err := config.DefaultDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}
	session := map[string]any{
		"token":     "token",
		"uid":       "user@example.com",
		"client":    "client",
		"cached_at": time.Now(),
	}
	b, _ := json.Marshal(session)
	if err := os.WriteFile(filepath.Join(configDir, "session.json"), b, 0600); err != nil {
		t.Fatal(err)
	}
}
