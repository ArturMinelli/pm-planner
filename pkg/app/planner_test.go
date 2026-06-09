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
	if payload.Out2 == "" {
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
	if payload.Balance != nil || payload.BalanceError == "" || payload.Out2 == "" {
		t.Fatalf("payload should retain normal plan with balance error: %#v", payload)
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
	got, err := RecalculatePlanner("2026-06-09", 8.5*60*60, &balance, "08:00", "12:00", "13:30")
	if err != nil {
		t.Fatal(err)
	}
	if got.Out2 != "18:00" || got.AlternativeOut2 != "19:00" {
		t.Fatalf("clockouts: %#v", got)
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RecalculatePlanner("2026-06-09", 8.5*60*60, &tt.balance, "08:00", "12:00", "13:30")
			if err != nil {
				t.Fatal(err)
			}
			if got.Out2 != "18:00" || got.AlternativeOut2 != tt.alternative {
				t.Fatalf("clockouts: %#v", got)
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
	got, err := RecalculatePlanner("2026-06-08", 8.5*60*60, &balance, "08:00", "12:00", "13:30")
	if err != nil {
		t.Fatal(err)
	}
	if got.Out2 != "18:00" || got.AlternativeOut2 != "" {
		t.Fatalf("non-today clockouts: %#v", got)
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
