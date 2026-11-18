package punchlog

import (
	"context"
	"sort"
	"sync"
	"time"

	"pm-cli/pkg/api"
)

const (
	SyncInterval     = 3 * time.Hour
	PendingInterval  = 5 * time.Minute
	UnconfirmedAfter = 5 * time.Minute
	// MatchWindow is how far a remote punch may be from a local record's time.
	MatchWindow = 3 * time.Minute
	// Debounce is the minimum gap between on-demand syncs.
	Debounce = 10 * time.Second
)

// Remote fetches the punches of today (satisfied by *api.Client).
type Remote interface {
	FetchTodayPunches(ctx context.Context) ([]api.Punch, error)
}

// State is what the UI shows. Counts are global, not per date.
type State struct {
	Date             string     `json:"date"`
	Punches          []Record   `json:"punches"`
	LastSyncAt       *time.Time `json:"lastSyncAt,omitempty"`
	LastSyncError    string     `json:"lastSyncError,omitempty"`
	PendingCount     int        `json:"pendingCount"`
	UnconfirmedCount int        `json:"unconfirmedCount"`
}

// Syncer reconciles the store with the remote. Exported fields may be set before use.
type Syncer struct {
	Store  *Store
	Remote Remote

	Now              func() time.Time
	SyncEvery        time.Duration
	PendingEvery     time.Duration
	UnconfirmedAfter time.Duration
	// NewTicker is replaceable in tests; stop must be safe to call once.
	NewTicker func(d time.Duration) (c <-chan time.Time, stop func())
	// OnUnconfirmed is called for each record that just became unconfirmed.
	OnUnconfirmed func(Record)

	syncMu       sync.Mutex // held for the duration of a remote sync
	lastFinished time.Time  // guarded by syncMu
}

// NewSyncer builds a Syncer with the default intervals and real clock.
func NewSyncer(store *Store, remote Remote) *Syncer {
	return &Syncer{
		Store: store, Remote: remote, Now: time.Now,
		SyncEvery: SyncInterval, PendingEvery: PendingInterval, UnconfirmedAfter: UnconfirmedAfter,
		NewTicker: func(d time.Duration) (<-chan time.Time, func()) {
			t := time.NewTicker(d)
			return t.C, t.Stop
		},
	}
}

// State returns the local state for date without any remote call. It still promotes
// overdue pending records to unconfirmed so the warning never lags the 5 min check tick.
func (s *Syncer) State(date string) (State, error) {
	var promoted []Record
	now := s.Now()
	if err := s.Store.Update(func(d *Data) bool {
		promoted = promote(d, now, s.UnconfirmedAfter)
		return len(promoted) > 0
	}); err != nil {
		return State{}, err
	}
	s.notify(promoted)
	d, err := s.Store.Read()
	if err != nil {
		return State{}, err
	}
	st := State{Date: date, Punches: []Record{}, LastSyncAt: d.LastSyncAt, LastSyncError: d.LastSyncError}
	for _, r := range d.Records {
		switch r.Status {
		case StatusPending:
			st.PendingCount++
		case StatusUnconfirmed:
			st.UnconfirmedCount++
		}
		if r.Date == date {
			st.Punches = append(st.Punches, r)
		}
	}
	sort.SliceStable(st.Punches, func(i, j int) bool { return st.Punches[i].Time < st.Punches[j].Time })
	return st, nil
}

// Sync runs a full sync now (blocking behind any sync in flight).
func (s *Syncer) Sync(ctx context.Context) error {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()
	return s.syncLocked(ctx)
}

// SyncOnDemand is the debounced entry for user-triggered syncs: if one is in flight or
// finished less than Debounce ago it makes no remote call.
func (s *Syncer) SyncOnDemand(ctx context.Context) (State, error) {
	if s.syncMu.TryLock() {
		if s.lastFinished.IsZero() || s.Now().Sub(s.lastFinished) >= Debounce {
			_ = s.syncLocked(ctx) // the error is surfaced through State.LastSyncError
		}
		s.syncMu.Unlock()
	}
	return s.State(s.Now().Format("2006-01-02"))
}

// CheckPending re-checks only when pending or unconfirmed records exist.
func (s *Syncer) CheckPending(ctx context.Context) error {
	d, err := s.Store.Read()
	if err != nil {
		return err
	}
	for _, r := range d.Records {
		if r.Status == StatusPending || r.Status == StatusUnconfirmed {
			return s.Sync(ctx)
		}
	}
	return nil
}

func (s *Syncer) syncLocked(ctx context.Context) error {
	remote, fetchErr := s.Remote.FetchTodayPunches(ctx)
	now := s.Now()
	s.lastFinished = now
	var promoted []Record
	err := s.Store.Update(func(d *Data) bool {
		if fetchErr != nil {
			// Keep every record as is (auth/network failures prove nothing about a punch).
			d.LastSyncError = fetchErr.Error()
		} else {
			promoted = reconcile(d, remote, now, s.UnconfirmedAfter)
			d.LastSyncAt = &now
			d.LastSyncError = ""
		}
		promoted = append(promoted, promote(d, now, s.UnconfirmedAfter)...)
		return true
	})
	s.notify(promoted)
	if fetchErr != nil {
		return fetchErr
	}
	return err
}

// ReconcileDay reconciles open records against the punches of one fetched work day
// (each carrying its date), with the same matching and dedupe as a regular sync.
func (s *Syncer) ReconcileDay(remote []api.Punch) error {
	if len(remote) == 0 {
		return nil
	}
	now := s.Now()
	var promoted []Record
	err := s.Store.Update(func(d *Data) bool {
		promoted = reconcile(d, remote, now, s.UnconfirmedAfter)
		return true
	})
	s.notify(promoted)
	return err
}

func (s *Syncer) notify(rs []Record) {
	if s.OnUnconfirmed != nil {
		for _, r := range rs {
			s.OnUnconfirmed(r)
		}
	}
}

// Run syncs at start, fully every SyncEvery, and re-checks pending records every PendingEvery.
func (s *Syncer) Run(ctx context.Context) {
	syncC, stopSync := s.NewTicker(s.SyncEvery)
	defer stopSync()
	pendC, stopPend := s.NewTicker(s.PendingEvery)
	defer stopPend()
	_ = s.Sync(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-syncC:
			_ = s.Sync(ctx)
		case <-pendC:
			_ = s.CheckPending(ctx)
		}
	}
}

// promote flips pending records older than after to unconfirmed and returns them.
func promote(d *Data, now time.Time, after time.Duration) []Record {
	var out []Record
	for i := range d.Records {
		r := &d.Records[i]
		if r.Status == StatusPending && now.Sub(r.CreatedAt) >= after {
			r.Status = StatusUnconfirmed
			out = append(out, *r)
		}
	}
	return out
}

func parseTime(date, tm string, loc *time.Location) (time.Time, bool) {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", date+" "+tm, loc)
	if err != nil { // work_days returns HH:MM
		t, err = time.ParseInLocation("2006-01-02 15:04", date+" "+tm, loc)
	}
	return t, err == nil
}

// reconcile merges the remote punches of today into d. Pending/unconfirmed records claim
// the nearest unclaimed active remote punch within MatchWindow; remote punches nobody
// claims are added as source "remote". A remoteId never appears in two records.
// Matching is by the record's own date, not today: the remote may still return an earlier
// day's punches after midnight, and a record left open overnight must still link. An open
// record also absorbs an already mirrored remote record of its day (stale duplicates).
// shortcut: days the remote no longer returns are not re-checked, FetchWorkDay has no ids.
func reconcile(d *Data, remote []api.Punch, now time.Time, after time.Duration) []Record {
	absorbMirrored(d, now)
	byRemote := map[int64]*Record{}
	for i := range d.Records {
		if id := d.Records[i].RemoteID; id != 0 {
			byRemote[id] = &d.Records[i]
		}
	}
	// Known punches: mirror the disabled flag.
	for _, p := range remote {
		if r := byRemote[p.ID]; r != nil {
			r.Disabled = p.Disabled
		}
	}
	// Open records (oldest first) claim the nearest free active punch.
	var open []*Record
	for i := range d.Records {
		r := &d.Records[i]
		if r.RemoteID == 0 && (r.Status == StatusPending || r.Status == StatusUnconfirmed) {
			open = append(open, r)
		}
	}
	sort.SliceStable(open, func(i, j int) bool { return open[i].CreatedAt.Before(open[j].CreatedAt) })
	for _, r := range open {
		rt, ok := parseTime(r.Date, r.Time, now.Location())
		if !ok {
			continue
		}
		r.Checks++
		r.LastCheckedAt = &now
		best := -1
		var bestDiff time.Duration
		for i, p := range remote {
			if p.Disabled || p.ID == 0 || byRemote[p.ID] != nil || p.Date != r.Date {
				continue
			}
			pt, ok := parseTime(p.Date, p.Time, now.Location())
			if !ok {
				continue
			}
			diff := pt.Sub(rt)
			if diff < 0 {
				diff = -diff
			}
			if diff <= MatchWindow && (best < 0 || diff < bestDiff) {
				best, bestDiff = i, diff
			}
		}
		if best >= 0 {
			p := remote[best]
			r.RemoteID, r.Status, r.SyncedAt = p.ID, StatusSynced, &now
			r.Date, r.Time = p.Date, p.Time
			byRemote[p.ID] = r
		}
	}
	// Remaining remote punches are new to us.
	for _, p := range remote {
		if p.ID == 0 || byRemote[p.ID] != nil {
			continue
		}
		d.Records = append(d.Records, Record{
			LocalID: NewLocalID(), Date: p.Date, Time: p.Time, CreatedAt: now,
			Source: SourceRemote, Status: StatusSynced, RemoteID: p.ID, Address: p.Address,
			Latitude: p.Latitude, Longitude: p.Longitude, Disabled: p.Disabled, SyncedAt: &now,
		})
		byRemote[p.ID] = &d.Records[len(d.Records)-1]
	}
	return promote(d, now, after)
}

// absorbMirrored links each open record to the nearest remote-sourced record of the same
// date within MatchWindow and drops that mirror, so the punch shows once.
func absorbMirrored(d *Data, now time.Time) {
	drop := map[int]bool{}
	for i := range d.Records {
		r := &d.Records[i]
		if r.RemoteID != 0 || (r.Status != StatusPending && r.Status != StatusUnconfirmed) {
			continue
		}
		rt, ok := parseTime(r.Date, r.Time, now.Location())
		if !ok {
			continue
		}
		best, bestDiff := -1, MatchWindow+1
		for j := range d.Records {
			m := &d.Records[j]
			if drop[j] || m.Source != SourceRemote || m.Disabled || m.RemoteID == 0 || m.Date != r.Date {
				continue
			}
			mt, ok := parseTime(m.Date, m.Time, now.Location())
			if !ok {
				continue
			}
			diff := mt.Sub(rt)
			if diff < 0 {
				diff = -diff
			}
			if diff < bestDiff {
				best, bestDiff = j, diff
			}
		}
		if best >= 0 {
			r.RemoteID, r.Status, r.SyncedAt, r.Time = d.Records[best].RemoteID, StatusSynced, &now, d.Records[best].Time
			drop[best] = true
		}
	}
	kept := d.Records[:0]
	for i, r := range d.Records {
		if !drop[i] {
			kept = append(kept, r)
		}
	}
	d.Records = kept
}
