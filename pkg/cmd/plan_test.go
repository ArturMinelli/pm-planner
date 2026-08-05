package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/plan"
)

func TestFormatAlternativeTime(t *testing.T) {
	balance := &plan.BalanceAdjustment{
		BalanceSecs:          -5400,
		AppliesToday:         true,
		TargetAdjustmentSecs: 3600,
	}
	got := formatAlternativeTime("19:00", balance)
	if got != "Horário alternativo: 19:00 (banco -01:30)" {
		t.Fatalf("alternative line: %q", got)
	}

	balance.AppliesToday = false
	if got := formatAlternativeTime("19:00", balance); got != "" {
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

func TestLoadPlanPayloadFallsBackToDefaultsOnFetchFailure(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, ".config"))
	if err := os.MkdirAll(filepath.Join(dir, ".config", "pm"), 0700); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	client := &api.Client{
		HTTP:    http.DefaultClient,
		BaseURL: srv.URL,
		AuthHeaders: func() (map[string]string, error) {
			return map[string]string{"Access-Token": "test"}, nil
		},
	}

	payload, warning, err := loadPlanPayload(context.Background(), client, "2026-06-09")
	if err != nil {
		t.Fatal(err)
	}
	if warning != planDefaultsWarning {
		t.Fatalf("warning: %q", warning)
	}
	if payload.BaseTargetSecs != int64((8*time.Hour + 30*time.Minute).Seconds()) {
		t.Fatalf("defaults target: %d", payload.BaseTargetSecs)
	}
	if len(payload.Journeys) != 2 || !payload.SolvedSlot.Valid() || payload.SolvedSlot.JourneyIndex != 1 {
		t.Fatalf("defaults shape: %#v", payload)
	}
	if payload.OriginalsLine != "(nenhum)" {
		t.Fatalf("originals: %q", payload.OriginalsLine)
	}
	if payload.Balance != nil {
		t.Fatalf("defaults should omit balance: %#v", payload.Balance)
	}
}

func TestLoadPlanPayloadUsesAPIWhenAvailable(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, ".config"))
	configDir := filepath.Join(dir, ".config", "pm")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}
	session := `{"token":"token","uid":"user@example.com","client":"client","cached_at":"` + time.Now().Format(time.RFC3339) + `"}`
	if err := os.WriteFile(filepath.Join(configDir, "session.json"), []byte(session), 0600); err != nil {
		t.Fatal(err)
	}

	today := "2026-06-09"
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

	client := &api.Client{
		HTTP:    http.DefaultClient,
		BaseURL: srv.URL,
		AuthHeaders: func() (map[string]string, error) {
			return map[string]string{"Access-Token": "test"}, nil
		},
	}

	payload, warning, err := loadPlanPayload(context.Background(), client, today)
	if err != nil {
		t.Fatal(err)
	}
	if warning != "" {
		t.Fatalf("unexpected warning: %q", warning)
	}
	if payload.BaseTargetSecs != 30600 {
		t.Fatalf("api target: %d", payload.BaseTargetSecs)
	}
}
