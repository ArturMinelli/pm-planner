package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pm-cli/pkg/config"
)

func TestFetchClockInStatusReturnsLatestPunch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/time_cards/current/last_cached" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"time_cards":[{"id":1,"date":"2026-06-17","time":"08:01:07","address":"Rua A, 1","latitude":-27.6,"longitude":-48.6}]}`))
	}))
	defer srv.Close()

	got, err := timecardTestClient(srv.URL).FetchClockInStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !got.HasRecent || got.Date != "2026-06-17" || got.Time != "08:01:07" || got.Address != "Rua A, 1" {
		t.Fatalf("status: %#v", got)
	}
}

func TestRegisterTimeCardBuildsPayloadAndParsesResponse(t *testing.T) {
	setupTimecardSession(t)
	const (
		testLat = -27.123456
		testLng = -48.654321
	)
	var gotPayload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/employees/current":
			_, _ = w.Write([]byte(`{"employee":{"id":99,"login":"user@example.com","name":"Test User"}}`))
		case "/time_cards/register":
			if r.Header.Get("Api-Version") != "2" {
				t.Errorf("Api-Version header: got %q", r.Header.Get("Api-Version"))
			}
			if r.Header.Get("Token-Type") != "Bearer" {
				t.Errorf("Token-Type header: got %q", r.Header.Get("Token-Type"))
			}
			if r.Header.Get("Expiry") == "" {
				t.Error("Expiry header missing")
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(body, &gotPayload); err != nil {
				t.Fatal(err)
			}
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"untreated_time_card":{"time":"11:45","date":"2026-06-17","address":"Rua A, 1"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	got, err := timecardTestClient(srv.URL).RegisterTimeCard(context.Background(), TimeCardLocation{
		Latitude:  testLat,
		Longitude: testLng,
		Address:   "Rua A, 1",
		Accuracy:  25,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Time != "11:45" || got.Date != "2026-06-17" || got.Address != "Rua A, 1" {
		t.Fatalf("result: %#v", got)
	}

	employee, ok := gotPayload["employee"].(map[string]any)
	if !ok || employee["id"] != float64(99) {
		t.Fatalf("employee payload: %#v", gotPayload["employee"])
	}
	timeCard, ok := gotPayload["time_card"].(map[string]any)
	if !ok {
		t.Fatal("time_card missing")
	}
	if timeCard["latitude"] != testLat || timeCard["longitude"] != testLng || timeCard["address"] != "Rua A, 1" {
		t.Fatalf("time_card payload: %#v", timeCard)
	}
	if gotPayload["_path"] != "/registrar-ponto" {
		t.Fatalf("_path: %#v", gotPayload["_path"])
	}
	device, ok := gotPayload["_device"].(map[string]any)
	if !ok {
		t.Fatal("_device missing")
	}
	uuidObj, ok := device["uuid"].(map[string]any)
	if !ok || uuidObj["token"] == nil {
		t.Fatalf("_device.uuid: %#v", device["uuid"])
	}
}

func TestRegisterTimeCardRejectsNonSuccess(t *testing.T) {
	setupTimecardSession(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/employees/current":
			_, _ = w.Write([]byte(`{"employee":{"id":99}}`))
		case "/time_cards/register":
			http.Error(w, "forbidden", http.StatusForbidden)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	_, err := timecardTestClient(srv.URL).RegisterTimeCard(context.Background(), TimeCardLocation{
		Latitude:  -27.1,
		Longitude: -48.1,
		Address:   "Rua A",
		Accuracy:  10,
	})
	if err == nil {
		t.Fatal("expected register error")
	}
}

func TestNewAuthenticatedRequestEncodesJSONBody(t *testing.T) {
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("content-type: %q", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := timecardTestClient(srv.URL)
	req, err := client.NewAuthenticatedRequest(context.Background(), http.MethodPost, srv.URL, map[string]string{"hello": "world"})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.HTTP.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if !strings.Contains(body, `"hello":"world"`) {
		t.Fatalf("body: %s", body)
	}
}

func timecardTestClient(baseURL string) *Client {
	return &Client{
		HTTP:    http.DefaultClient,
		BaseURL: baseURL,
		AuthHeaders: func() (map[string]string, error) {
			return map[string]string{
				"Access-Token": "test",
				"Token":        "test",
				"Uid":          "user@example.com",
				"Client":       "client",
				"Api-Version":  "2",
				"uuid":         "device-uuid-1",
				"Token-Type":   "Bearer",
				"Expiry":       "2000000000",
			}, nil
		},
	}
}

func setupTimecardSession(t *testing.T) {
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
		"token":            "test",
		"uid":              "user@example.com",
		"client":           "client",
		"token_type":       "Bearer",
		"expiry":           int64(2000000000),
		"uuid":             "device-uuid-1",
		"sign_in_success":  "ok",
		"sign_in_count":    1,
		"last_sign_in_ip":  "127.0.0.1",
		"last_sign_in_at":  1,
		"cached_at":        time.Now(),
	}
	b, _ := json.Marshal(session)
	if err := os.WriteFile(filepath.Join(path, "session.json"), b, 0600); err != nil {
		t.Fatal(err)
	}
}
