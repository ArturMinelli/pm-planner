package server

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/punchlog"
)

var (
	punchMu     sync.Mutex
	punchSyncer *punchlog.Syncer
)

// PunchLog returns the process-wide syncer shared by clock-in, HTTP, Wails and the scheduler.
func PunchLog() *punchlog.Syncer {
	punchMu.Lock()
	defer punchMu.Unlock()
	if punchSyncer == nil {
		s := punchlog.NewSyncer(punchlog.DefaultStore(), api.New())
		// Dev-only overrides to watch pending -> unconfirmed without waiting minutes.
		envDuration("PM_PUNCH_PENDING_INTERVAL", &s.PendingEvery)
		envDuration("PM_PUNCH_UNCONFIRMED_AFTER", &s.UnconfirmedAfter)
		punchSyncer = s
	}
	return punchSyncer
}

// SetPunchLog replaces the shared syncer (tests); nil resets to lazy default.
func SetPunchLog(s *punchlog.Syncer) {
	punchMu.Lock()
	punchSyncer = s
	punchMu.Unlock()
}

func envDuration(name string, dst *time.Duration) {
	if d, err := time.ParseDuration(os.Getenv(name)); err == nil && d > 0 {
		*dst = d
	}
}

// GetPunchLog returns the local registry for date (default today); no remote call.
func GetPunchLog(date string) (punchlog.State, error) {
	s := PunchLog()
	if date == "" {
		date = s.Now().Format("2006-01-02")
	} else if _, err := time.Parse("2006-01-02", date); err != nil {
		return punchlog.State{}, fmt.Errorf("invalid date %q", date)
	}
	return s.State(date)
}

// SyncPunches runs a debounced on-demand sync and returns today's state.
func SyncPunches(ctx context.Context) (punchlog.State, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return PunchLog().SyncOnDemand(reqCtx)
}
