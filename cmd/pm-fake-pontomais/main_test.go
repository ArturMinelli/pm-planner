package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pm-cli/pkg/api"
)

// Synthetic fixture only.
const validRegister = `{"employee":{"id":1},"time_card":{"latitude":-22.9,"longitude":-47.0,"address":"FAKE+00 Test City"}}`

type punchResp struct {
	TimeCard punch  `json:"time_card"`
	Message  string `json:"message_id"`
}

type callsResp struct {
	RegisterCalls int     `json:"registerCalls"`
	Punches       []punch `json:"punches"`
}

func call(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
}

func expectCode(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d (body %s)", rec.Code, want, rec.Body.String())
	}
}

func TestOKFlow(t *testing.T) {
	h := newHandler()

	rec := call(t, h, http.MethodOptions, "/api/session", "")
	expectCode(t, rec, http.StatusNoContent)
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("missing permissive CORS header")
	}

	var sess struct {
		Session struct {
			Employee struct {
				ID int `json:"id"`
			} `json:"employee"`
		} `json:"session"`
	}
	rec = call(t, h, http.MethodGet, "/api/session", "")
	expectCode(t, rec, http.StatusOK)
	decode(t, rec, &sess)
	if sess.Session.Employee.ID != 1 {
		t.Fatalf("employee id = %d, want 1", sess.Session.Employee.ID)
	}

	rec = call(t, h, http.MethodGet, "/api/geo_locations?lat=-22.9&lng=-47&employee_id=1", "")
	expectCode(t, rec, http.StatusOK)
	if !strings.Contains(rec.Body.String(), "FAKE+00 Test City") {
		t.Fatalf("geo body missing address: %s", rec.Body.String())
	}

	rec = call(t, h, http.MethodPost, "/api/time_cards/register", `{"employee":{"id":1}}`)
	expectCode(t, rec, http.StatusUnprocessableEntity)

	rec = call(t, h, http.MethodPost, "/api/time_cards/register", validRegister)
	expectCode(t, rec, http.StatusAccepted)
	var pr punchResp
	decode(t, rec, &pr)
	if pr.TimeCard.ID != 9000001 || pr.Message != "fake" || pr.TimeCard.Disabled {
		t.Fatalf("unexpected register response: %+v", pr)
	}

	var last struct {
		TimeCards []punch `json:"time_cards"`
	}
	rec = call(t, h, http.MethodGet, "/api/time_cards/current/last_cached", "")
	expectCode(t, rec, http.StatusOK)
	decode(t, rec, &last)
	if len(last.TimeCards) != 1 || last.TimeCards[0].ID != 9000001 {
		t.Fatalf("last_cached = %+v, want one punch 9000001", last.TimeCards)
	}
}

func TestUnauthorizedDoesNotStore(t *testing.T) {
	h := newHandler()
	expectCode(t, call(t, h, http.MethodPost, "/__scenario", `{"mode":"401"}`), http.StatusOK)
	expectCode(t, call(t, h, http.MethodPost, "/api/time_cards/register", validRegister), http.StatusUnauthorized)

	var calls callsResp
	decode(t, call(t, h, http.MethodGet, "/__calls", ""), &calls)
	if calls.RegisterCalls != 1 || len(calls.Punches) != 0 {
		t.Fatalf("calls = %+v, want 1 call and no punches", calls)
	}
}

func TestCounterIncrements(t *testing.T) {
	h := newHandler()
	for i := 0; i < 2; i++ {
		expectCode(t, call(t, h, http.MethodPost, "/api/time_cards/register", validRegister), http.StatusAccepted)
	}
	var calls callsResp
	decode(t, call(t, h, http.MethodGet, "/__calls", ""), &calls)
	if calls.RegisterCalls != 2 || len(calls.Punches) != 2 {
		t.Fatalf("calls = %+v, want 2 calls and 2 punches", calls)
	}
	if calls.Punches[1].ID != 9000002 {
		t.Fatalf("second punch id = %d, want 9000002", calls.Punches[1].ID)
	}
}

func TestReset(t *testing.T) {
	h := newHandler()
	call(t, h, http.MethodPost, "/__scenario", `{"mode":"401"}`)
	call(t, h, http.MethodPost, "/api/time_cards/register", validRegister)
	expectCode(t, call(t, h, http.MethodPost, "/__reset", ""), http.StatusOK)

	var calls callsResp
	decode(t, call(t, h, http.MethodGet, "/__calls", ""), &calls)
	if calls.RegisterCalls != 0 || len(calls.Punches) != 0 {
		t.Fatalf("after reset calls = %+v, want zeros", calls)
	}
	// Scenario must be back to ok.
	expectCode(t, call(t, h, http.MethodPost, "/api/time_cards/register", validRegister), http.StatusAccepted)
}

func TestRegisterOnFailureStoresPunch(t *testing.T) {
	cases := []struct {
		name     string
		scenario string
		wantRows int
	}{
		{"500 without flag", `{"mode":"500"}`, 0},
		{"500 with flag", `{"mode":"500","registerOnFailure":true}`, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newHandler()
			expectCode(t, call(t, h, http.MethodPost, "/__scenario", tc.scenario), http.StatusOK)
			expectCode(t, call(t, h, http.MethodPost, "/api/time_cards/register", validRegister), http.StatusInternalServerError)

			var calls callsResp
			decode(t, call(t, h, http.MethodGet, "/__calls", ""), &calls)
			if calls.RegisterCalls != 1 || len(calls.Punches) != tc.wantRows {
				t.Fatalf("calls = %+v, want 1 call and %d punches", calls, tc.wantRows)
			}
		})
	}
}

func TestWorkDayReflectsRegisteredPunch(t *testing.T) {
	h := newHandler()
	today := time.Now().Format("2006-01-02")
	expectCode(t, call(t, h, http.MethodPost, "/api/time_cards/register", validRegister), http.StatusAccepted)

	rec := call(t, h, http.MethodGet, "/api/time_card_control/current/work_days/"+today, "")
	expectCode(t, rec, http.StatusOK)
	var env struct {
		WorkDay api.WorkDay `json:"work_day"`
	}
	decode(t, rec, &env)
	wd := env.WorkDay
	if wd.Date != today || wd.ShiftTime != 28800 || len(wd.ShiftDay.Periods) != 2 {
		t.Fatalf("work day = %+v", wd)
	}
	if len(wd.TimeCards) != 1 {
		t.Fatalf("time cards = %+v, want one punch", wd.TimeCards)
	}
	if _, err := time.Parse("15:04", wd.TimeCards[0].Time); err != nil {
		t.Fatalf("time card %q not HH:MM: %v", wd.TimeCards[0].Time, err)
	}

	// A different date has no punches but still a schedule.
	var other struct {
		WorkDay api.WorkDay `json:"work_day"`
	}
	decode(t, call(t, h, http.MethodGet, "/api/time_card_control/current/work_days/1999-01-01", ""), &other)
	if other.WorkDay.Date != "1999-01-01" || len(other.WorkDay.TimeCards) != 0 {
		t.Fatalf("other day = %+v, want no time cards", other.WorkDay)
	}
}

func TestBalanceEndpointsParse(t *testing.T) {
	h := newHandler()
	var emp struct {
		Employee struct {
			ID json.Number `json:"id"`
		} `json:"employee"`
	}
	rec := call(t, h, http.MethodGet, "/api/employees/my_time_break", "")
	expectCode(t, rec, http.StatusOK)
	decode(t, rec, &emp)
	if emp.Employee.ID.String() != "1" {
		t.Fatalf("employee id = %q, want 1", emp.Employee.ID)
	}

	var st struct {
		Statuses struct {
			TimeBalance int64 `json:"time_balance"`
		} `json:"statuses"`
	}
	rec = call(t, h, http.MethodGet, "/api/employees/statuses/1", "")
	expectCode(t, rec, http.StatusOK)
	decode(t, rec, &st)
}
