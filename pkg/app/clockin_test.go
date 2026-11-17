package app

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"pm-cli/pkg/api"
)

type fetchStep struct {
	punches []api.Punch
	err     error
}

// fakePuncher returns fetches in order (the last one repeats) and records register calls.
type fakePuncher struct {
	mu          sync.Mutex
	fetches     []fetchStep
	fetchN      int
	employeeErr error
	geoErr      error
	regErr      error
	regPunch    *api.Punch
	delay       time.Duration
	registers   []api.Location
	inFlight    int
	maxInFlight int
}

func (f *fakePuncher) FetchEmployee(ctx context.Context) (json.RawMessage, string, error) {
	if f.employeeErr != nil {
		return nil, "", f.employeeErr
	}
	return json.RawMessage(`{"id":42}`), "42", nil
}

func (f *fakePuncher) ReverseGeocode(ctx context.Context, lat, lng float64, employeeID string) (string, error) {
	if f.geoErr != nil {
		return "", f.geoErr
	}
	return "Rua Teste, 1", nil
}

func (f *fakePuncher) RegisterPunch(ctx context.Context, loc api.Location) (*api.Punch, error) {
	f.mu.Lock()
	f.inFlight++
	f.maxInFlight = max(f.maxInFlight, f.inFlight)
	f.registers = append(f.registers, loc)
	f.mu.Unlock()

	time.Sleep(f.delay)

	f.mu.Lock()
	f.inFlight--
	f.mu.Unlock()
	if f.regErr != nil {
		return nil, f.regErr
	}
	return f.regPunch, nil
}

func (f *fakePuncher) FetchTodayPunches(ctx context.Context) ([]api.Punch, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.fetches) == 0 {
		f.fetchN++
		return nil, nil
	}
	i := min(f.fetchN, len(f.fetches)-1)
	f.fetchN++
	return f.fetches[i].punches, f.fetches[i].err
}

func (f *fakePuncher) registerCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.registers)
}

var testNow = time.Date(2026, 11, 17, 10, 0, 0, 0, time.Local)

func pinNow(t *testing.T, at time.Time) {
	t.Helper()
	orig := now
	now = func() time.Time { return at }
	t.Cleanup(func() { now = orig })
}

func validLoc() ClockInLocation {
	return ClockInLocation{Latitude: -23.55, Longitude: -46.63, Accuracy: 12, Address: "Av. Paulista, 1", Source: "detected"}
}

func TestClockInSuccess(t *testing.T) {
	pinNow(t, testNow)
	p := &fakePuncher{regPunch: &api.Punch{ID: 7, Date: "2026-11-17", Time: "09:00:00"}}

	res := ClockIn(context.Background(), p, ClockInRequest{Location: validLoc()})

	if res.Status != "registered" || res.Registered == nil || !*res.Registered {
		t.Fatalf("got %+v, want registered/true", res)
	}
	if res.Punch == nil || res.Punch.ID != 7 || res.LocationSource != "detected" {
		t.Fatalf("got punch %+v source %q", res.Punch, res.LocationSource)
	}
	if p.registerCount() != 1 {
		t.Fatalf("register calls = %d, want 1", p.registerCount())
	}
}

func TestClockInResolvesAddress(t *testing.T) {
	pinNow(t, testNow)
	p := &fakePuncher{regPunch: &api.Punch{ID: 1}}
	loc := validLoc()
	loc.Address = ""

	ClockIn(context.Background(), p, ClockInRequest{Location: loc})
	if got := p.registers[0].Address; got != "Rua Teste, 1" {
		t.Fatalf("address = %q, want geocoded value", got)
	}

	p = &fakePuncher{regPunch: &api.Punch{ID: 1}, geoErr: errors.New("boom")}
	ClockIn(context.Background(), p, ClockInRequest{Location: loc})
	if got := p.registers[0].Address; got != "-23.550000, -46.630000" {
		t.Fatalf("address = %q, want coordinate fallback", got)
	}
}

func TestClockInPrefetchNetworkErrorContinues(t *testing.T) {
	pinNow(t, testNow)
	p := &fakePuncher{
		fetches:  []fetchStep{{err: errors.New("dial tcp: refused")}},
		regPunch: &api.Punch{ID: 2},
	}
	res := ClockIn(context.Background(), p, ClockInRequest{Location: validLoc()})
	if res.Status != "registered" {
		t.Fatalf("got %+v, want registered despite pre-fetch failure", res)
	}
}

func TestClockInInvalidLocation(t *testing.T) {
	pinNow(t, testNow)
	cases := map[string]ClockInLocation{
		"lat too high": {Latitude: 91, Longitude: 10, Source: "detected"},
		"lng too low":  {Latitude: 10, Longitude: -181, Source: "detected"},
		"both zero":    {Latitude: 0, Longitude: 0, Source: "default"},
	}
	for name, loc := range cases {
		t.Run(name, func(t *testing.T) {
			p := &fakePuncher{regPunch: &api.Punch{ID: 1}}
			res := ClockIn(context.Background(), p, ClockInRequest{Location: loc})
			if res.Status != "failed" || res.Code != "invalid_location" || res.Registered == nil || *res.Registered {
				t.Fatalf("got %+v, want failed/invalid_location/false", res)
			}
			if p.registerCount() != 0 {
				t.Fatalf("sent a punch for invalid location")
			}
		})
	}
}

func TestClockInTooCloseBlocked(t *testing.T) {
	pinNow(t, testNow)
	p := &fakePuncher{fetches: []fetchStep{{punches: []api.Punch{
		{ID: 5, Date: "2026-11-17", Time: "09:59:30"},
	}}}}

	res := ClockIn(context.Background(), p, ClockInRequest{Location: validLoc()})
	if res.Status != "failed" || res.Code != "too_close" || res.Registered == nil || *res.Registered {
		t.Fatalf("got %+v, want failed/too_close/false", res)
	}
	if p.registerCount() != 0 {
		t.Fatalf("punch sent despite too_close")
	}
}

func TestClockInTooCloseIgnoresDisabled(t *testing.T) {
	pinNow(t, testNow)
	p := &fakePuncher{
		fetches:  []fetchStep{{punches: []api.Punch{{ID: 5, Date: "2026-11-17", Time: "09:59:30", Disabled: true}}}},
		regPunch: &api.Punch{ID: 6},
	}
	res := ClockIn(context.Background(), p, ClockInRequest{Location: validLoc()})
	if res.Status != "registered" {
		t.Fatalf("got %+v, want registered (disabled punch must not trigger guard)", res)
	}
}

func TestClockInTooCloseConfirmedSends(t *testing.T) {
	pinNow(t, testNow)
	p := &fakePuncher{
		fetches:  []fetchStep{{punches: []api.Punch{{ID: 5, Date: "2026-11-17", Time: "09:59:30"}}}},
		regPunch: &api.Punch{ID: 6},
	}
	res := ClockIn(context.Background(), p, ClockInRequest{Location: validLoc(), ConfirmNear: true})
	if res.Status != "registered" || p.registerCount() != 1 {
		t.Fatalf("got %+v with %d registers, want registered with 1 send", res, p.registerCount())
	}
}

func TestClockInSessionExpired(t *testing.T) {
	pinNow(t, testNow)
	p := &fakePuncher{regErr: &api.HTTPStatusError{StatusCode: 401, Body: "nope"}}

	res := ClockIn(context.Background(), p, ClockInRequest{Location: validLoc()})
	if res.Status != "failed" || res.Code != "session_expired" || res.Registered == nil || *res.Registered {
		t.Fatalf("got %+v, want failed/session_expired/false", res)
	}
	if p.fetchN != 1 {
		t.Fatalf("fetch calls = %d, want 1 (no verify after 4xx)", p.fetchN)
	}
}

func TestClockInServiceRejected(t *testing.T) {
	pinNow(t, testNow)
	p := &fakePuncher{regErr: &api.HTTPStatusError{StatusCode: 422, Body: "bad"}}

	res := ClockIn(context.Background(), p, ClockInRequest{Location: validLoc()})
	if res.Status != "failed" || res.Code != "service_rejected" || res.Registered == nil || *res.Registered {
		t.Fatalf("got %+v, want failed/service_rejected/false", res)
	}
	if p.fetchN != 1 {
		t.Fatalf("fetch calls = %d, want 1 (no verify after 4xx)", p.fetchN)
	}
}

func TestClockInServerErrorRecovered(t *testing.T) {
	pinNow(t, testNow)
	p := &fakePuncher{
		regErr: &api.HTTPStatusError{StatusCode: 500, Body: "oops"},
		fetches: []fetchStep{
			{punches: nil},
			{punches: []api.Punch{{ID: 9, Date: "2026-11-17", Time: "10:00:00"}}},
		},
	}

	res := ClockIn(context.Background(), p, ClockInRequest{Location: validLoc()})
	if res.Status != "registered" || res.Code != "recovered" || res.Registered == nil || !*res.Registered {
		t.Fatalf("got %+v, want registered/recovered/true", res)
	}
	if res.Punch == nil || res.Punch.ID != 9 {
		t.Fatalf("punch = %+v, want ID 9", res.Punch)
	}
}

func TestClockInNetworkErrorNoPunch(t *testing.T) {
	pinNow(t, testNow)
	p := &fakePuncher{
		regErr: errors.New("i/o timeout"),
		fetches: []fetchStep{
			{punches: nil},
			{punches: nil},
		},
	}

	res := ClockIn(context.Background(), p, ClockInRequest{Location: validLoc()})
	if res.Status != "failed" || res.Code != "offline" || res.Registered == nil || *res.Registered {
		t.Fatalf("got %+v, want failed/offline/false", res)
	}
}

func TestClockInUncertainWhenVerifyFails(t *testing.T) {
	pinNow(t, testNow)
	p := &fakePuncher{
		regErr: errors.New("context deadline exceeded"),
		fetches: []fetchStep{
			{punches: nil},
			{err: errors.New("still offline")},
		},
	}

	res := ClockIn(context.Background(), p, ClockInRequest{Location: validLoc()})
	if res.Status != "uncertain" || res.Code != "uncertain" || res.Registered != nil {
		t.Fatalf("got %+v, want uncertain with nil Registered", res)
	}
}

func TestClockInConcurrentCallsSerialized(t *testing.T) {
	pinNow(t, testNow)
	p := &fakePuncher{regPunch: &api.Punch{ID: 1}, delay: 30 * time.Millisecond}

	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ClockIn(context.Background(), p, ClockInRequest{Location: validLoc(), ConfirmNear: true})
		}()
	}
	wg.Wait()

	if p.registerCount() != 2 {
		t.Fatalf("register calls = %d, want 2", p.registerCount())
	}
	if p.maxInFlight != 1 {
		t.Fatalf("max concurrent registers = %d, want 1", p.maxInFlight)
	}
}
