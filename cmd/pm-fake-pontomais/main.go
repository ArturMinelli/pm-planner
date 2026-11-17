// Command pm-fake-pontomais is a localhost-only stand-in for the PontoMais API,
// used so UI tests never touch production. Fixtures are synthetic.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type punch struct {
	ID        int     `json:"id"`
	Date      string  `json:"date"`
	Time      string  `json:"time"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Disabled  bool    `json:"disabled"`
}

type scenario struct {
	Mode              string `json:"mode"`
	RegisterOnFailure bool   `json:"registerOnFailure"`
}

var validModes = map[string]bool{"ok": true, "401": true, "500": true, "slow": true, "drop": true}

type fake struct {
	mu      sync.Mutex
	punches []punch
	calls   int
	nextID  int
	sc      scenario
}

func (f *fake) reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.punches = []punch{}
	f.calls = 0
	f.nextID = 9000001
	f.sc = scenario{Mode: "ok"}
}

func newHandler() http.Handler {
	f := &fake{}
	f.reset()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/session", f.session)
	mux.HandleFunc("GET /api/geo_locations", f.geo)
	mux.HandleFunc("POST /api/time_cards/register", f.register)
	mux.HandleFunc("GET /api/time_cards/current/last_cached", f.lastCached)
	mux.HandleFunc("GET /api/time_card_control/current/work_days/{date}", f.workDay)
	mux.HandleFunc("GET /api/employees/my_time_break", f.myTimeBreak)
	mux.HandleFunc("GET /api/employees/statuses/{id}", f.statuses)
	mux.HandleFunc("POST /__scenario", f.setScenario)
	mux.HandleFunc("GET /__calls", f.callsInfo)
	mux.HandleFunc("POST /__reset", func(w http.ResponseWriter, r *http.Request) {
		f.reset()
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
	return logAndCORS(mux)
}

// logAndCORS logs method and path only (never bodies) and answers preflight.
func logAndCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(os.Stdout, r.Method, r.URL.Path)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (f *fake) session(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"session": map[string]any{
			"employee": map[string]any{"id": 1, "name": "Test User", "login": "test@example.invalid"},
		},
		"meta": map[string]any{},
	})
}

func (f *fake) geo(w http.ResponseWriter, r *http.Request) {
	lat, err1 := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, err2 := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	if err1 != nil || err2 != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid lat/lng"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"address": map[string]any{"formatted_address": "FAKE+00 Test City", "latitude": lat, "longitude": lng},
		"meta":    map[string]any{},
	})
}

func (f *fake) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Employee struct {
			ID *int `json:"id"`
		} `json:"employee"`
		TimeCard *struct {
			Latitude  *float64 `json:"latitude"`
			Longitude *float64 `json:"longitude"`
			Address   *string  `json:"address"`
		} `json:"time_card"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Employee.ID == nil ||
		body.TimeCard == nil || body.TimeCard.Latitude == nil || body.TimeCard.Longitude == nil || body.TimeCard.Address == nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "invalid body"})
		return
	}

	f.mu.Lock()
	f.calls++
	sc := f.sc
	// Stored before the outcome is applied when the scenario is "uncertain" or succeeds.
	var p punch
	if sc.Mode == "ok" || sc.Mode == "slow" || (sc.RegisterOnFailure && (sc.Mode == "500" || sc.Mode == "drop")) {
		now := time.Now()
		p = punch{
			ID:        f.nextID,
			Date:      now.Format("2006-01-02"),
			Time:      now.Format("15:04:05"),
			Address:   *body.TimeCard.Address,
			Latitude:  *body.TimeCard.Latitude,
			Longitude: *body.TimeCard.Longitude,
		}
		f.nextID++
		f.punches = append(f.punches, p)
	}
	f.mu.Unlock()

	switch sc.Mode {
	case "401":
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	case "500":
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal"})
	case "drop":
		hj, ok := w.(http.Hijacker)
		if !ok {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "cannot hijack"})
			return
		}
		if conn, _, err := hj.Hijack(); err == nil {
			_ = conn.Close()
		}
	default:
		if sc.Mode == "slow" {
			time.Sleep(12 * time.Second)
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"time_card": p, "message_id": "fake"})
	}
}

func (f *fake) lastCached(w http.ResponseWriter, _ *http.Request) {
	f.mu.Lock()
	list := make([]punch, 0, len(f.punches))
	for i := len(f.punches) - 1; i >= 0; i-- {
		list = append(list, f.punches[i])
	}
	f.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{
		"time_cards":      list,
		"register_alerts": []any{},
		"meta":            map[string]any{},
	})
}

// workDay serves a fixed synthetic schedule plus the registered punches for {date}.
func (f *fake) workDay(w http.ResponseWriter, r *http.Request) {
	date := r.PathValue("date")
	cards := []map[string]string{}
	f.mu.Lock()
	for _, p := range f.punches {
		if p.Date == date {
			cards = append(cards, map[string]string{"time": p.Time[:5]}) // stored as HH:MM:SS; planner expects HH:MM
		}
	}
	f.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{
		"work_day": map[string]any{
			"date":       date,
			"shift_time": 28800,
			"shift_day": map[string]any{"periods": []map[string]string{
				{"enter_time": "09:00", "leave_time": "12:00"},
				{"enter_time": "13:00", "leave_time": "18:00"},
			}},
			"time_cards": cards,
		},
	})
}

func (f *fake) myTimeBreak(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"employee": map[string]any{"id": 1},
		"meta":     map[string]any{},
	})
}

func (f *fake) statuses(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"statuses": map[string]any{"time_balance": 0},
		"meta":     map[string]any{},
	})
}

func (f *fake) setScenario(w http.ResponseWriter, r *http.Request) {
	var sc scenario
	if err := json.NewDecoder(r.Body).Decode(&sc); err != nil || !validModes[sc.Mode] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid scenario"})
		return
	}
	f.mu.Lock()
	f.sc = sc
	f.mu.Unlock()
	writeJSON(w, http.StatusOK, sc)
}

func (f *fake) callsInfo(w http.ResponseWriter, _ *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"registerCalls": f.calls, "punches": f.punches})
}

func main() {
	port := os.Getenv("PM_FAKE_PORT")
	if port == "" {
		port = "3850"
	}
	ln, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", port))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stdout, "pm-fake-pontomais listening on", ln.Addr())
	log.Fatal(http.Serve(ln, newHandler()))
}
