package punchlog

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"pm-cli/pkg/api"
)

var t0 = time.Date(2026, 11, 17, 10, 0, 0, 0, time.UTC)

type fakeRemote struct {
	mu      sync.Mutex
	punches []api.Punch
	err     error
	calls   int
	called  chan struct{}
}

func (f *fakeRemote) FetchTodayPunches(ctx context.Context) ([]api.Punch, error) {
	f.mu.Lock()
	f.calls++
	p, err := f.punches, f.err
	f.mu.Unlock()
	if f.called != nil {
		f.called <- struct{}{}
	}
	return p, err
}

func (f *fakeRemote) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func newSyncer(t *testing.T, r *fakeRemote) (*Syncer, *clock) {
	t.Helper()
	c := &clock{t0}
	s := NewSyncer(NewStore(filepath.Join(t.TempDir(), "punches.json")), r)
	s.Now = c.now
	return s, c
}

func pendingAt(t *testing.T, s *Syncer, tm string, created time.Time) string {
	t.Helper()
	id, err := s.Store.AddPending(Record{Date: "2026-11-17", Time: tm, CreatedAt: created})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func byID(t *testing.T, s *Syncer, id string) Record {
	t.Helper()
	d, _ := s.Store.Read()
	for _, r := range d.Records {
		if r.LocalID == id {
			return r
		}
	}
	t.Fatalf("record %s missing", id)
	return Record{}
}

func rp(id int64, tm string, disabled bool) api.Punch {
	return api.Punch{ID: id, Date: "2026-11-17", Time: tm, Disabled: disabled}
}

func TestStoreRoundTripAndPerms(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "punches.json")
	st := NewStore(path)
	id, err := st.AddPending(Record{Date: "2026-11-17", Time: "09:00:00", CreatedAt: t0, Latitude: 1})
	if err != nil {
		t.Fatal(err)
	}
	d, err := NewStore(path).Read()
	if err != nil || len(d.Records) != 1 || d.Records[0].LocalID != id || d.Records[0].Status != StatusPending || d.Records[0].Source != SourceApp {
		t.Fatalf("round trip: %+v %v", d, err)
	}
	fi, _ := os.Stat(path)
	if fi.Mode().Perm() != 0600 {
		t.Fatalf("perm = %v", fi.Mode().Perm())
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatalf("temp files left behind: %v", entries)
	}
}

func TestStoreMissingAndCorrupt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "punches.json")
	st := NewStore(path)
	if d, err := st.Read(); err != nil || len(d.Records) != 0 {
		t.Fatalf("missing: %+v %v", d, err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0600); err != nil {
		t.Fatal(err)
	}
	if d, err := st.Read(); err != nil || len(d.Records) != 0 {
		t.Fatalf("corrupt: %+v %v", d, err)
	}
	if b, err := os.ReadFile(path + ".bak"); err != nil || string(b) != "{not json" {
		t.Fatalf("bak: %q %v", b, err)
	}
	if _, err := st.AddPending(Record{Date: "2026-11-17", Time: "09:00:00", CreatedAt: t0}); err != nil {
		t.Fatal(err)
	}
	if _, err := NewStore("").Read(); err == nil {
		t.Fatal("empty path must error")
	}
}

func TestReconcileMatching(t *testing.T) {
	tests := []struct {
		name       string
		recTime    string
		remote     []api.Punch
		wantRemote int64 // 0 = stays unmatched
	}{
		{"exact", "09:00:00", []api.Punch{rp(1, "09:00:00", false)}, 1},
		{"edge -3m", "09:00:00", []api.Punch{rp(1, "08:57:00", false)}, 1},
		{"edge +3m", "09:00:00", []api.Punch{rp(1, "09:03:00", false)}, 1},
		{"just outside", "09:00:00", []api.Punch{rp(1, "09:03:01", false)}, 0},
		{"nearest wins", "09:00:00", []api.Punch{rp(1, "09:02:00", false), rp(2, "08:59:30", false)}, 2},
		{"disabled ignored", "09:00:00", []api.Punch{rp(1, "09:00:00", true)}, 0},
		{"disabled skipped for active", "09:00:00", []api.Punch{rp(1, "09:00:00", true), rp(2, "09:01:00", false)}, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &fakeRemote{punches: tc.remote}
			s, _ := newSyncer(t, r)
			id := pendingAt(t, s, tc.recTime, t0)
			if err := s.Sync(context.Background()); err != nil {
				t.Fatal(err)
			}
			rec := byID(t, s, id)
			if rec.RemoteID != tc.wantRemote {
				t.Fatalf("remoteId = %d, want %d", rec.RemoteID, tc.wantRemote)
			}
			if tc.wantRemote != 0 && (rec.Status != StatusSynced || rec.SyncedAt == nil) {
				t.Fatalf("not synced: %+v", rec)
			}
			if tc.wantRemote == 0 && rec.Status != StatusPending {
				t.Fatalf("status = %s", rec.Status)
			}
			// No duplicate remoteIds ever.
			d, _ := s.Store.Read()
			seen := map[int64]bool{}
			for _, x := range d.Records {
				if x.RemoteID != 0 && seen[x.RemoteID] {
					t.Fatalf("duplicate remoteId %d", x.RemoteID)
				}
				seen[x.RemoteID] = true
			}
		})
	}
}

func TestReconcileClaimedAndRemoteOnly(t *testing.T) {
	r := &fakeRemote{punches: []api.Punch{rp(1, "09:00:00", false)}}
	s, _ := newSyncer(t, r)
	a := pendingAt(t, s, "09:00:00", t0)
	b := pendingAt(t, s, "09:00:30", t0.Add(time.Second))
	_ = s.Sync(context.Background())
	if byID(t, s, a).RemoteID != 1 || byID(t, s, b).RemoteID != 0 {
		t.Fatalf("one punch must be claimed once: a=%+v b=%+v", byID(t, s, a), byID(t, s, b))
	}
	// A punch already linked is not reused and not duplicated across syncs.
	_ = s.Sync(context.Background())
	_ = s.Sync(context.Background())
	d, _ := s.Store.Read()
	if len(d.Records) != 2 {
		t.Fatalf("records = %d, want 2", len(d.Records))
	}
	// Remote-only punch is added once with source remote.
	r.punches = append(r.punches, rp(5, "12:00:00", false))
	_ = s.Sync(context.Background())
	_ = s.Sync(context.Background())
	d, _ = s.Store.Read()
	var n int
	for _, x := range d.Records {
		if x.RemoteID == 5 {
			n++
			if x.Source != SourceRemote || x.Status != StatusSynced {
				t.Fatalf("remote-only: %+v", x)
			}
		}
	}
	if n != 1 {
		t.Fatalf("remote-only count = %d", n)
	}
}

func TestReconcileDisabledReflected(t *testing.T) {
	r := &fakeRemote{punches: []api.Punch{rp(1, "09:00:00", false)}}
	s, _ := newSyncer(t, r)
	_ = s.Sync(context.Background())
	r.punches = []api.Punch{rp(1, "09:00:00", true)}
	_ = s.Sync(context.Background())
	d, _ := s.Store.Read()
	if len(d.Records) != 1 || !d.Records[0].Disabled {
		t.Fatalf("%+v", d.Records)
	}
	r.punches = []api.Punch{rp(1, "09:00:00", false)}
	_ = s.Sync(context.Background())
	d, _ = s.Store.Read()
	if d.Records[0].Disabled {
		t.Fatal("re-enabled punch should clear disabled")
	}
}

func TestUnconfirmedAtExactlyFiveMinutesThenSynced(t *testing.T) {
	r := &fakeRemote{}
	s, c := newSyncer(t, r)
	var notified []Record
	s.OnUnconfirmed = func(x Record) { notified = append(notified, x) }
	id := pendingAt(t, s, "10:00:00", t0)

	c.t = t0.Add(UnconfirmedAfter - time.Second)
	_ = s.CheckPending(context.Background())
	if got := byID(t, s, id); got.Status != StatusPending || got.Checks != 1 {
		t.Fatalf("before 5m: %+v", got)
	}
	c.t = t0.Add(UnconfirmedAfter)
	_ = s.CheckPending(context.Background())
	if got := byID(t, s, id).Status; got != StatusUnconfirmed {
		t.Fatalf("at 5m: %s", got)
	}
	_ = s.CheckPending(context.Background())
	if len(notified) != 1 {
		t.Fatalf("notified %d times, want 1", len(notified))
	}
	st, _ := s.State("2026-11-17")
	if st.UnconfirmedCount != 1 || st.PendingCount != 0 {
		t.Fatalf("counts %+v", st)
	}

	r.punches = []api.Punch{rp(9, "10:01:00", false)}
	c.t = t0.Add(20 * time.Minute)
	_ = s.CheckPending(context.Background())
	got := byID(t, s, id)
	if got.Status != StatusSynced || got.RemoteID != 9 {
		t.Fatalf("late sync: %+v", got)
	}
	// Nothing open -> CheckPending makes no remote call.
	before := r.count()
	_ = s.CheckPending(context.Background())
	if r.count() != before {
		t.Fatal("CheckPending hit the remote with nothing open")
	}
}

func TestStatePromotesWithoutRemoteCall(t *testing.T) {
	r := &fakeRemote{}
	s, c := newSyncer(t, r)
	pendingAt(t, s, "10:00:00", t0)
	c.t = t0.Add(UnconfirmedAfter)
	st, err := s.State("2026-11-17")
	if err != nil || st.UnconfirmedCount != 1 || r.count() != 0 {
		t.Fatalf("%+v %v calls=%d", st, err, r.count())
	}
}

func TestSyncFailureKeepsRecords(t *testing.T) {
	r := &fakeRemote{punches: []api.Punch{rp(1, "09:00:00", false)}}
	s, c := newSyncer(t, r)
	id := pendingAt(t, s, "09:00:00", t0)
	_ = s.Sync(context.Background())
	pend := pendingAt(t, s, "11:00:00", t0)

	r.err = errors.New("HTTP 401: session expired")
	c.t = t0.Add(time.Hour)
	if err := s.Sync(context.Background()); err == nil {
		t.Fatal("want error")
	}
	st, _ := s.State("2026-11-17")
	if st.LastSyncError == "" || st.LastSyncAt == nil || len(st.Punches) != 2 {
		t.Fatalf("%+v", st)
	}
	if byID(t, s, id).Status != StatusSynced || byID(t, s, pend).Status != StatusUnconfirmed {
		t.Fatal("auth failure must not mark records failed or drop them")
	}
	r.err = nil
	_ = s.Sync(context.Background())
	if st, _ = s.State("2026-11-17"); st.LastSyncError != "" {
		t.Fatalf("error not cleared: %q", st.LastSyncError)
	}
}

func TestStateSortedGlobalCounts(t *testing.T) {
	s, _ := newSyncer(t, &fakeRemote{})
	_, _ = s.Store.AddPending(Record{Date: "2026-11-17", Time: "12:00:00", CreatedAt: t0})
	_, _ = s.Store.AddPending(Record{Date: "2026-11-17", Time: "08:00:00", CreatedAt: t0})
	_, _ = s.Store.AddPending(Record{Date: "2026-11-10", Time: "08:00:00", CreatedAt: t0})
	st, _ := s.State("2026-11-17")
	if len(st.Punches) != 2 || st.Punches[0].Time != "08:00:00" || st.PendingCount != 3 {
		t.Fatalf("%+v", st)
	}
	if st, _ = s.State("2020-01-01"); st.Punches == nil || len(st.Punches) != 0 || st.PendingCount != 3 {
		t.Fatalf("other day: %+v", st)
	}
}

func TestMarkSyncedDropsRemoteDuplicate(t *testing.T) {
	s, c := newSyncer(t, &fakeRemote{punches: []api.Punch{rp(7, "09:00:00", false)}})
	_ = s.Sync(context.Background()) // imports 7 as remote
	id := pendingAt(t, s, "09:00:10", t0)
	if err := s.Store.MarkSynced(id, rp(7, "09:00:00", false), c.t); err != nil {
		t.Fatal(err)
	}
	d, _ := s.Store.Read()
	if len(d.Records) != 1 || d.Records[0].LocalID != id || d.Records[0].Source != SourceApp {
		t.Fatalf("%+v", d.Records)
	}
}

func TestSyncOnDemandDebounce(t *testing.T) {
	r := &fakeRemote{}
	s, c := newSyncer(t, r)
	ctx := context.Background()
	_, _ = s.SyncOnDemand(ctx)
	c.t = t0.Add(Debounce - time.Second)
	_, _ = s.SyncOnDemand(ctx)
	if r.count() != 1 {
		t.Fatalf("calls = %d, want 1 (debounced)", r.count())
	}
	c.t = t0.Add(Debounce)
	_, _ = s.SyncOnDemand(ctx)
	if r.count() != 2 {
		t.Fatalf("calls = %d, want 2", r.count())
	}
}

func TestSyncOnDemandInFlight(t *testing.T) {
	r := &fakeRemote{called: make(chan struct{})}
	s, _ := newSyncer(t, r)
	done := make(chan struct{})
	go func() { _ = s.Sync(context.Background()); close(done) }()
	<-r.called // remote call is now blocked on the unbuffered channel, sync in flight
	st, err := s.SyncOnDemand(context.Background())
	if err != nil || st.Date == "" {
		t.Fatalf("%+v %v", st, err)
	}
	if r.count() != 1 {
		t.Fatalf("calls = %d, want 1", r.count())
	}
	<-done
}

func TestRunSchedule(t *testing.T) {
	r := &fakeRemote{called: make(chan struct{}, 16)}
	s, _ := newSyncer(t, r)
	syncC, pendC := make(chan time.Time), make(chan time.Time)
	got := map[time.Duration]bool{}
	s.NewTicker = func(d time.Duration) (<-chan time.Time, func()) {
		got[d] = true
		if d == SyncInterval {
			return syncC, func() {}
		}
		return pendC, func() {}
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { s.Run(ctx); close(done) }()
	wait := func() {
		select {
		case <-r.called:
		case <-time.After(2 * time.Second):
			t.Fatal("expected a remote call")
		}
	}
	wait() // startup sync

	pendC <- t0 // nothing open: no call
	syncC <- t0 // full sync: call
	wait()
	pendingAt(t, s, "10:00:00", t0)
	pendC <- t0 // open record: call
	wait()
	cancel()
	<-done
	if r.count() != 3 || !got[SyncInterval] || !got[PendingInterval] {
		t.Fatalf("calls=%d tickers=%v", r.count(), got)
	}
}
