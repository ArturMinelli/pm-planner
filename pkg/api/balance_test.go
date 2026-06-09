package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pm-cli/pkg/auth"
	"pm-cli/pkg/config"
)

func TestFetchEmployeeBalanceDiscoversAndParsesNumericSeconds(t *testing.T) {
	setupBalanceSession(t, "")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/employees/my_time_break":
			_, _ = w.Write([]byte(`{"employee":{"id":2751858}}`))
		case "/employees/statuses/2751858":
			_, _ = w.Write([]byte(`{"statuses":{"time_balance":-5400,"last_settle_time_balance":{"updated_at":"2026-06-09T12:00:00-03:00"}}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	got, err := balanceTestClient(srv.URL).FetchEmployeeBalance(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.EmployeeID != "2751858" || got.TimeBalanceSecs != -5400 || got.UpdatedAt == "" {
		t.Fatalf("balance: %#v", got)
	}
	if cached := auth.GetCachedEmployeeID(); cached != "2751858" {
		t.Fatalf("cached employee ID: got %q", cached)
	}
}

func TestFetchEmployeeBalanceParsesStringSeconds(t *testing.T) {
	setupBalanceSession(t, "42")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"statuses":{"time_balance":"3600","last_settle_time_balance":null}}`))
	}))
	defer srv.Close()

	got, err := balanceTestClient(srv.URL).FetchEmployeeBalance(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.TimeBalanceSecs != 3600 {
		t.Fatalf("seconds: got %d", got.TimeBalanceSecs)
	}
}

func TestFetchEmployeeBalanceParsesDecimalWholeSeconds(t *testing.T) {
	for _, value := range []string{`-31560.0`, `"-31560.0"`} {
		t.Run(value, func(t *testing.T) {
			setupBalanceSession(t, "42")
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"statuses":{"time_balance":` + value + `}}`))
			}))
			defer srv.Close()

			got, err := balanceTestClient(srv.URL).FetchEmployeeBalance(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if got.TimeBalanceSecs != -31560 {
				t.Fatalf("seconds: got %d", got.TimeBalanceSecs)
			}
		})
	}
}

func TestFetchEmployeeBalanceRejectsFractionalSeconds(t *testing.T) {
	setupBalanceSession(t, "42")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"statuses":{"time_balance":1.5}}`))
	}))
	defer srv.Close()

	if _, err := balanceTestClient(srv.URL).FetchEmployeeBalance(context.Background()); err == nil {
		t.Fatal("expected fractional seconds error")
	}
}

func TestFetchEmployeeBalanceReturnsNonSuccessStatus(t *testing.T) {
	setupBalanceSession(t, "42")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := balanceTestClient(srv.URL).FetchEmployeeBalance(context.Background())
	statusErr, ok := err.(*HTTPStatusError)
	if !ok || statusErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("error: %#v", err)
	}
}

func TestFetchEmployeeBalanceRediscoversStaleCachedID(t *testing.T) {
	setupBalanceSession(t, "old")
	var discoveryCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/employees/statuses/old":
			http.Error(w, "stale", http.StatusNotFound)
		case "/employees/my_time_break":
			discoveryCalls++
			_, _ = w.Write([]byte(`{"employee":{"id":"new"}}`))
		case "/employees/statuses/new":
			_, _ = w.Write([]byte(`{"statuses":{"time_balance":0}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	got, err := balanceTestClient(srv.URL).FetchEmployeeBalance(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.EmployeeID != "new" || discoveryCalls != 1 || auth.GetCachedEmployeeID() != "new" {
		t.Fatalf("rediscovery result: %#v calls=%d cached=%q", got, discoveryCalls, auth.GetCachedEmployeeID())
	}
}

func balanceTestClient(baseURL string) *Client {
	return &Client{
		HTTP:    http.DefaultClient,
		BaseURL: baseURL,
		AuthHeaders: func() (map[string]string, error) {
			return map[string]string{"Access-Token": "test"}, nil
		},
	}
}

func setupBalanceSession(t *testing.T, employeeID string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, ".config"))
	path, err := config.DefaultDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(path, 0700); err != nil {
		t.Fatal(err)
	}
	session := map[string]any{
		"token":       "token",
		"uid":         "user@example.com",
		"client":      "client",
		"cached_at":   time.Now(),
		"employee_id": employeeID,
	}
	b, _ := json.Marshal(session)
	if err := os.WriteFile(filepath.Join(path, "session.json"), b, 0600); err != nil {
		t.Fatal(err)
	}
}
