package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/spf13/viper"
	"pm-cli/pkg/auth"
)

const syntheticEmployee = `{"id":999001,"name":"Synthetic Employee","company_id":1}`

func punchTestClient(srv *httptest.Server) *Client {
	return &Client{
		HTTP:    srv.Client(),
		BaseURL: srv.URL + "/api",
		AuthHeaders: func() (map[string]string, error) {
			return map[string]string{"uid": "x"}, nil
		},
	}
}

func sessionHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(`{"session":{"employee":` + syntheticEmployee + `}}`))
}

func TestFetchEmployee(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/session" {
			http.NotFound(w, r)
			return
		}
		sessionHandler(w, r)
	}))
	defer srv.Close()

	raw, id, err := punchTestClient(srv).FetchEmployee(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if id != "999001" {
		t.Fatalf("employee id: got %q", id)
	}
	var got, want map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	_ = json.Unmarshal([]byte(syntheticEmployee), &want)
	if got["name"] != want["name"] || got["company_id"] != want["company_id"] {
		t.Fatalf("employee raw: got %s", raw)
	}
}

func TestReverseGeocode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/geo_locations" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		if q.Get("lat") != "-27.5" || q.Get("lng") != "-48.5" || q.Get("employee_id") != "999001" {
			t.Errorf("query: %v", q)
		}
		_, _ = w.Write([]byte(`{"address":{"formatted_address":"SYNTH+AB Testville"}}`))
	}))
	defer srv.Close()

	got, err := punchTestClient(srv).ReverseGeocode(context.Background(), -27.5, -48.5, "999001")
	if err != nil {
		t.Fatal(err)
	}
	if got != "SYNTH+AB Testville" {
		t.Fatalf("address: got %q", got)
	}
}

func TestFetchTodayPunches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/time_cards/current/last_cached" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"time_cards":[{"id":7,"date":"2026-01-02","time":"08:00","address":"Synthetic St","latitude":1.5,"longitude":2.5,"disabled":false},{"id":8,"date":"2026-01-02","time":"12:00","address":"Synthetic St","latitude":1.5,"longitude":2.5,"disabled":true}]}`))
	}))
	defer srv.Close()

	got, err := punchTestClient(srv).FetchTodayPunches(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("punches: got %d", len(got))
	}
	if got[0].ID != 7 || got[0].Time != "08:00" || got[0].Latitude != 1.5 || got[0].Disabled {
		t.Fatalf("punch 0: %+v", got[0])
	}
	if got[1].ID != 8 || !got[1].Disabled {
		t.Fatalf("punch 1: %+v", got[1])
	}
}

var uuidV4 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestRegisterPunchAccepted(t *testing.T) {
	var gotBody map[string]any
	var gotUUID, gotAPIVersion, gotUID, gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/session":
			sessionHandler(w, r)
		case "/api/time_cards/register":
			if r.Method != http.MethodPost {
				t.Errorf("method: %s", r.Method)
			}
			gotUUID = r.Header.Get("uuid")
			gotAPIVersion = r.Header.Get("api-version")
			gotUID = r.Header.Get("uid")
			gotCT = r.Header.Get("Content-Type")
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Errorf("decode body: %v", err)
			}
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"time_card":{"id":4242,"date":"2026-01-02","time":"08:01","address":"SYNTH+AB Testville","latitude":-27.5,"longitude":-48.5,"disabled":false},"message_id":"m1"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	punch, err := punchTestClient(srv).RegisterPunch(context.Background(), Location{
		Latitude: -27.5, Longitude: -48.5, Accuracy: 1100, Address: "SYNTH+AB Testville",
	})
	if err != nil {
		t.Fatal(err)
	}
	if punch.ID != 4242 || punch.Time != "08:01" || punch.Address != "SYNTH+AB Testville" {
		t.Fatalf("punch: %+v", punch)
	}
	if gotAPIVersion != "2" || gotUID != "x" || gotCT != "application/json" {
		t.Fatalf("headers: api-version=%q uid=%q content-type=%q", gotAPIVersion, gotUID, gotCT)
	}
	if !uuidV4.MatchString(gotUUID) {
		t.Fatalf("uuid header not v4: %q", gotUUID)
	}

	if gotBody["_path"] != "/registrar-ponto" || gotBody["_appVersion"] != "0.10.32" {
		t.Fatalf("body meta: %v", gotBody)
	}
	dev, _ := gotBody["_device"].(map[string]any)
	if dev["uuid"] != gotUUID {
		t.Fatalf("_device.uuid %v != uuid header %q", dev["uuid"], gotUUID)
	}
	emp, _ := gotBody["employee"].(map[string]any)
	if emp["id"] != float64(999001) {
		t.Fatalf("employee not embedded: %v", gotBody["employee"])
	}
	tc, _ := gotBody["time_card"].(map[string]any)
	checks := map[string]any{
		"latitude": -27.5, "longitude": -48.5, "address": "SYNTH+AB Testville",
		"original_latitude": -27.5, "original_longitude": -48.5, "original_address": "SYNTH+AB Testville",
		"accuracy": float64(1100), "location_edited": false,
	}
	for k, want := range checks {
		if tc[k] != want {
			t.Errorf("time_card.%s: got %v want %v", k, tc[k], want)
		}
	}
	if v, ok := gotBody["image"]; !ok || v != nil {
		t.Errorf("image should be null: %v", v)
	}
}

func TestRegisterPunchAcceptedWithoutTimeCard(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/session" {
			sessionHandler(w, r)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = io.WriteString(w, `{"message_id":"m1"}`)
	}))
	defer srv.Close()

	punch, err := punchTestClient(srv).RegisterPunch(context.Background(), Location{Latitude: 1, Longitude: 2})
	if err != nil {
		t.Fatal(err)
	}
	if punch == nil || punch.ID != 0 {
		t.Fatalf("expected zero-ID punch, got %+v", punch)
	}
}

func TestRegisterPunch500IsHTTPStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/session" {
			sessionHandler(w, r)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":"synthetic"}`)
	}))
	defer srv.Close()

	_, err := punchTestClient(srv).RegisterPunch(context.Background(), Location{})
	var statusErr *HTTPStatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected *HTTPStatusError, got %T: %v", err, err)
	}
	if statusErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d", statusErr.StatusCode)
	}
}

func TestRegisterPunch401IsHTTPStatusError(t *testing.T) {
	// A 401 makes DoAuthenticated re-login and retry once. Keep that on a local sign-in
	// server with synthetic credentials so no real host is contacted.
	setupBalanceSession(t, "")
	oldSignInURL := auth.SignInURL
	oldLogin := viper.GetString("login")
	oldPassword := viper.GetString("password")
	t.Cleanup(func() {
		auth.SignInURL = oldSignInURL
		viper.Set("login", oldLogin)
		viper.Set("password", oldPassword)
	})
	viper.Set("login", "synthetic@example.com")
	viper.Set("password", "secret")
	signInSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("access-token", "relogin-token")
		w.Header().Set("client", "relogin-client")
		w.Header().Set("uid", "synthetic@example.com")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token":     "relogin-token",
			"client_id": "relogin-client",
			"data":      map[string]any{"login": "synthetic@example.com"},
		})
	}))
	defer signInSrv.Close()
	auth.SignInURL = signInSrv.URL

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/session" {
			sessionHandler(w, r)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":"synthetic"}`)
	}))
	defer srv.Close()

	client := &Client{HTTP: srv.Client(), BaseURL: srv.URL + "/api"}
	_, err := client.RegisterPunch(context.Background(), Location{})
	var statusErr *HTTPStatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected *HTTPStatusError, got %T: %v", err, err)
	}
	if statusErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d", statusErr.StatusCode)
	}
}
