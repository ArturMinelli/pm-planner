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
	if payload.BaseTargetSecs != 30600 || payload.TargetSecs != 34200 {
		t.Fatalf("targets: base=%d adjusted=%d", payload.BaseTargetSecs, payload.TargetSecs)
	}
	if payload.Balance == nil || payload.Balance.TargetAdjustmentSecs != 3600 {
		t.Fatalf("balance: %#v", payload.Balance)
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
	if payload.TargetSecs != payload.BaseTargetSecs || payload.Balance != nil || payload.BalanceError == "" {
		t.Fatalf("payload should retain normal plan with balance error: %#v", payload)
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
