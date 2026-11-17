package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/punchlog"
)

// Puncher is the subset of the PontoMais client used by ClockIn; *api.Client satisfies it.
type Puncher interface {
	FetchEmployee(ctx context.Context) (json.RawMessage, string, error)
	ReverseGeocode(ctx context.Context, lat, lng float64, employeeID string) (string, error)
	RegisterPunch(ctx context.Context, loc api.Location) (*api.Punch, error)
	FetchTodayPunches(ctx context.Context) ([]api.Punch, error)
}

// ClockInLocation is the position the user is clocking in from.
type ClockInLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Accuracy  float64 `json:"accuracy"`
	Address   string  `json:"address,omitempty"`
	Source    string  `json:"source"`
}

// ClockInRequest is one clock-in attempt. ConfirmNear overrides the recent-punch guard.
type ClockInRequest struct {
	Location    ClockInLocation `json:"location"`
	ConfirmNear bool            `json:"confirmNear"`
}

// ClockInResult reports the outcome. Registered is nil when it cannot be known.
type ClockInResult struct {
	Status         string     `json:"status"`
	Registered     *bool      `json:"registered"`
	Code           string     `json:"code,omitempty"`
	Punch          *api.Punch `json:"punch,omitempty"`
	LocationSource string     `json:"locationSource,omitempty"`
	// Detail is a short technical hint (step, HTTP status, truncated body) for failures.
	Detail string `json:"detail,omitempty"`
	// LocalID is the punch log record created for this attempt, if any.
	LocalID string `json:"localId,omitempty"`
}

// clockInMu serializes clock-in calls so a double click cannot send two punches (FR-002).
var clockInMu sync.Mutex

// now is a variable so tests can pin the clock.
var now = time.Now

const recentPunchGuard = 60 * time.Second

// Option customizes ClockIn.
type Option func(*options)

type options struct {
	log     *punchlog.Store
	localID string
}

// WithLog records the attempt in the local punch log (pending before sending).
func WithLog(s *punchlog.Store) Option { return func(o *options) { o.log = s } }

// ClockIn registers a punch. It never returns an error; the outcome is in the result.
func ClockIn(ctx context.Context, p Puncher, req ClockInRequest, opts ...Option) ClockInResult {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	res := clockIn(ctx, p, req, &o)
	res.LocalID = o.localID
	o.settle(res)
	return res
}

// begin saves the pending record right before the punch is sent. A log failure must not
// block the punch itself, so it only leaves the attempt untracked.
func (o *options) begin(loc ClockInLocation, address string) {
	if o.log == nil {
		return
	}
	t := now()
	id, err := o.log.AddPending(punchlog.Record{
		Date: t.Format("2006-01-02"), Time: t.Format("15:04:05"), CreatedAt: t,
		Address: address, Latitude: loc.Latitude, Longitude: loc.Longitude, LocationSource: loc.Source,
	})
	if err == nil {
		o.localID = id
	}
}

// settle maps the outcome onto the record; uncertain (or a registered punch without an
// id) stays pending for the syncer to confirm.
func (o *options) settle(res ClockInResult) {
	if o.log == nil || o.localID == "" {
		return
	}
	switch {
	case res.Status == "registered" && res.Punch != nil && res.Punch.ID != 0:
		_ = o.log.MarkSynced(o.localID, *res.Punch, now())
	case res.Status == "failed":
		_ = o.log.MarkFailed(o.localID)
	}
}

func clockIn(ctx context.Context, p Puncher, req ClockInRequest, o *options) ClockInResult {
	loc := req.Location
	if !(loc.Latitude >= -90 && loc.Latitude <= 90 && loc.Longitude >= -180 && loc.Longitude <= 180) ||
		(loc.Latitude == 0 && loc.Longitude == 0) {
		return ClockInResult{Status: "failed", Code: "invalid_location", Registered: boolPtr(false)}
	}

	clockInMu.Lock()
	defer clockInMu.Unlock()

	// A failed pre-fetch does not block the punch; the baseline is then unknown.
	before, baseErr := p.FetchTodayPunches(ctx)
	baseKnown := baseErr == nil
	if !baseKnown {
		before = nil
	}

	if !req.ConfirmNear {
		if last := newestPunch(before); last != nil {
			if t, _ := punchTime(*last); absDuration(now().Sub(t)) < recentPunchGuard {
				return ClockInResult{Status: "failed", Code: "too_close", Registered: boolPtr(false)}
			}
		}
	}

	address := loc.Address
	if address == "" {
		address = resolveAddress(ctx, p, loc)
	}

	o.begin(loc, address)
	punch, err := p.RegisterPunch(ctx, api.Location{
		Latitude:  loc.Latitude,
		Longitude: loc.Longitude,
		Accuracy:  loc.Accuracy,
		Address:   address,
	})
	if err == nil {
		return ClockInResult{Status: "registered", Registered: boolPtr(true), Punch: punch, LocationSource: loc.Source}
	}

	failCode := "offline"
	var statusErr *api.HTTPStatusError
	if errors.As(err, &statusErr) {
		switch {
		case statusErr.StatusCode == 401 || statusErr.StatusCode == 403:
			return ClockInResult{Status: "failed", Code: "session_expired", Registered: boolPtr(false), Detail: detail(err)}
		case statusErr.StatusCode >= 500:
			failCode = "service_rejected"
		default:
			return ClockInResult{Status: "failed", Code: "service_rejected", Registered: boolPtr(false), Detail: detail(err)}
		}
	}

	// Ambiguous: the service may have stored the punch despite the error. Check before reporting.
	after, verifyErr := p.FetchTodayPunches(ctx)
	if verifyErr != nil {
		return ClockInResult{Status: "uncertain", Code: "uncertain"}
	}
	if !baseKnown {
		// Without a baseline, any punch today may be an older one, so the outcome is not provable.
		if newestPunch(after) != nil {
			return ClockInResult{Status: "uncertain", Code: "uncertain"}
		}
	} else if pick := findNewPunch(before, after); pick != nil {
		return ClockInResult{Status: "registered", Registered: boolPtr(true), Code: "recovered", Punch: pick, LocationSource: loc.Source}
	}
	return ClockInResult{Status: "failed", Code: failCode, Registered: boolPtr(false), Detail: detail(err)}
}

// detail returns err as a short single-line hint for the UI.
func detail(err error) string {
	s := strings.Join(strings.Fields(err.Error()), " ")
	if len(s) > 200 {
		s = s[:200] + "…"
	}
	return s
}

// resolveAddress returns the service-formatted address, or raw coordinates if lookup fails.
func resolveAddress(ctx context.Context, p Puncher, loc ClockInLocation) string {
	fallback := fmt.Sprintf("%.6f, %.6f", loc.Latitude, loc.Longitude)
	_, employeeID, err := p.FetchEmployee(ctx)
	if err != nil {
		return fallback
	}
	address, err := p.ReverseGeocode(ctx, loc.Latitude, loc.Longitude, employeeID)
	if err != nil || address == "" {
		return fallback
	}
	return address
}

// findNewPunch returns the active punch in after that was not in before, or nil.
// If the active count grew without a new ID, it falls back to the newest active punch.
func findNewPunch(before, after []api.Punch) *api.Punch {
	seen := make(map[int64]bool, len(before))
	for _, b := range before {
		seen[b.ID] = true
	}
	var pick *api.Punch
	for i := range after {
		if !after[i].Disabled && !seen[after[i].ID] {
			pick = &after[i]
		}
	}
	if pick == nil && countActive(after) > countActive(before) {
		pick = newestPunch(after)
	}
	return pick
}

func countActive(list []api.Punch) int {
	n := 0
	for _, p := range list {
		if !p.Disabled {
			n++
		}
	}
	return n
}

// newestPunch returns a copy of the latest active punch with a parseable time, or nil.
func newestPunch(list []api.Punch) *api.Punch {
	var best *api.Punch
	var bestTime time.Time
	for i := range list {
		if list[i].Disabled {
			continue
		}
		t, ok := punchTime(list[i])
		if ok && (best == nil || t.After(bestTime)) {
			best, bestTime = &list[i], t
		}
	}
	if best == nil {
		return nil
	}
	c := *best
	return &c
}

// punchTime parses the service date and time as local time.
func punchTime(p api.Punch) (time.Time, bool) {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", p.Date+" "+p.Time, time.Local)
	return t, err == nil
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

func boolPtr(b bool) *bool { return &b }
