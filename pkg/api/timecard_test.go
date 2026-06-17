package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	const (
		testLat = -27.123456
		testLng = -48.654321
	)
	var gotPayload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/employees/my_time_break":
			_, _ = w.Write([]byte(`{"employee":{"id":99,"login":"user@example.com","name":"Test User"}}`))
		case "/time_cards/register":
			if r.Header.Get("Api-Version") != "2" {
				t.Errorf("Api-Version header: got %q", r.Header.Get("Api-Version"))
			}
			if r.Header.Get("uuid") != "device-uuid-1" {
				t.Errorf("uuid header: got %q", r.Header.Get("uuid"))
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
}

func TestRegisterTimeCardRejectsNonSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/employees/my_time_break":
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
	statusErr, ok := err.(*HTTPStatusError)
	if !ok || statusErr.StatusCode != http.StatusForbidden {
		t.Fatalf("error: %#v", err)
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
			}, nil
		},
	}
}
