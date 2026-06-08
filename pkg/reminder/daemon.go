package reminder

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/config"
)

const (
	defaultRefreshInterval     = time.Hour
	defaultFreshWindow         = 30 * time.Minute
	defaultPostSlotRefresh     = 2 * time.Minute
	defaultInitialBackoff      = time.Minute
	defaultMaxBackoff          = 30 * time.Minute
	defaultRequestTimeout      = 25 * time.Second
	defaultIdleDisabledRecheck = 5 * time.Minute
)

type Fetcher interface {
	FetchWorkDay(ctx context.Context, date string) (*api.WorkDay, error)
}

type Alerter interface {
	SendReminder(ctx context.Context, event ScheduledReminder, settings config.Reminders) error
}

type DaemonOptions struct {
	ConfigPath      string
	Fetcher         Fetcher
	Alerter         Alerter
	Store           *FileStore
	Logger          *log.Logger
	Now             func() time.Time
	RefreshInterval time.Duration
	FreshWindow     time.Duration
	PostSlotRefresh time.Duration
	RequestTimeout  time.Duration
	InitialBackoff  time.Duration
	MaxBackoff      time.Duration
	DisabledRecheck time.Duration
}

type Daemon struct {
	options DaemonOptions
}

type dayState struct {
	plan      *DayPlan
	fetchedAt time.Time
}

func NewDaemon(options DaemonOptions) *Daemon {
	if options.Logger == nil {
		options.Logger = log.New(os.Stderr, "pm-reminders: ", log.LstdFlags)
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.RefreshInterval <= 0 {
		options.RefreshInterval = defaultRefreshInterval
	}
	if options.FreshWindow <= 0 {
		options.FreshWindow = defaultFreshWindow
	}
	if options.PostSlotRefresh <= 0 {
		options.PostSlotRefresh = defaultPostSlotRefresh
	}
	if options.RequestTimeout <= 0 {
		options.RequestTimeout = defaultRequestTimeout
	}
	if options.InitialBackoff <= 0 {
		options.InitialBackoff = defaultInitialBackoff
	}
	if options.MaxBackoff <= 0 {
		options.MaxBackoff = defaultMaxBackoff
	}
	if options.DisabledRecheck <= 0 {
		options.DisabledRecheck = defaultIdleDisabledRecheck
	}
	return &Daemon{options: options}
}

func DefaultStatePath() (string, error) {
	dir, err := config.DefaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "reminders-state.json"), nil
}

func DefaultPIDPath() (string, error) {
	dir, err := config.DefaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "reminder-daemon.pid"), nil
}

func (d *Daemon) Run(ctx context.Context) error {
	if d.options.Fetcher == nil {
		return errors.New("fetcher is required")
	}
	if d.options.Alerter == nil {
		return errors.New("alerter is required")
	}
	store := d.options.Store
	if store == nil {
		path, err := DefaultStatePath()
		if err != nil {
			return err
		}
		store = NewFileStore(path)
		d.options.Store = store
	}

	var state *dayState
	var lastPostSlotRefresh time.Time
	backoff := d.options.InitialBackoff
	suppressed := map[string]bool{}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		settings, err := d.readSettings()
		if err != nil {
			d.logf("read settings: %v", err)
			if !d.sleep(ctx, jitter(backoff)) {
				return ctx.Err()
			}
			backoff = nextBackoff(backoff, d.options.MaxBackoff)
			continue
		}
		if !settings.Enabled {
			if !d.sleep(ctx, d.options.DisabledRecheck) {
				return ctx.Err()
			}
			continue
		}

		now := d.options.Now()
		today := now.Format("2006-01-02")
		_ = store.PruneBefore(today)
		if state == nil || state.plan == nil || state.plan.Date != today ||
			now.Sub(state.fetchedAt) >= d.options.RefreshInterval {
			nextState, wait, err := d.fetchDay(ctx, now)
			if err != nil {
				if wait <= d.options.InitialBackoff {
					wait = backoff
				}
				d.logf("refresh failed: %v", err)
				if !d.sleep(ctx, jitter(wait)) {
					return ctx.Err()
				}
				backoff = nextBackoff(wait, d.options.MaxBackoff)
				continue
			}
			state = nextState
			backoff = d.options.InitialBackoff
		}

		due := DueReminders(state.plan, settings, store, suppressed, now)
		if len(due) > 0 {
			verified, wait, err := d.fetchDay(ctx, now)
			if err != nil {
				if now.Sub(state.fetchedAt) > d.options.FreshWindow {
					for _, event := range due {
						suppressed[event.ID] = true
					}
					d.logf("suppressed %d stale reminders after verification failure: %v", len(due), err)
				} else {
					d.logf("verification failed; using fresh cached workday: %v", err)
				}
				if wait > 0 && wait < d.options.RefreshInterval {
					lastPostSlotRefresh = now.Add(wait)
				}
			} else {
				state = verified
				backoff = d.options.InitialBackoff
			}
			due = DueReminders(state.plan, settings, store, suppressed, d.options.Now())
			for _, event := range due {
				if err := d.options.Alerter.SendReminder(ctx, event, settings); err != nil {
					d.logf("send reminder %s: %v", event.ID, err)
					continue
				}
				if err := store.MarkDelivered(event.ID, d.options.Now()); err != nil {
					d.logf("mark delivered %s: %v", event.ID, err)
				}
				lastPostSlotRefresh = event.SlotTime.Add(d.options.PostSlotRefresh)
			}
			continue
		}

		wake, ok := NextWake(state.plan, settings, store, suppressed, now)
		nextRefresh := state.fetchedAt.Add(d.options.RefreshInterval)
		if lastPostSlotRefresh.After(now) && lastPostSlotRefresh.Before(nextRefresh) {
			nextRefresh = lastPostSlotRefresh
		}
		if !ok || nextRefresh.Before(wake) {
			wake = nextRefresh
		}
		if wake.Before(now) {
			wake = now.Add(time.Second)
		}
		if !d.sleep(ctx, wake.Sub(now)) {
			return ctx.Err()
		}
	}
}

func (d *Daemon) readSettings() (config.Reminders, error) {
	f, err := config.Read(d.options.ConfigPath)
	if err != nil {
		return config.Reminders{}, err
	}
	return config.ResolveReminders(f)
}

func (d *Daemon) fetchDay(ctx context.Context, now time.Time) (*dayState, time.Duration, error) {
	cfg, err := config.Read(d.options.ConfigPath)
	if err != nil {
		return nil, d.options.InitialBackoff, err
	}
	anchors, err := config.ResolvePlannerAnchors(cfg)
	if err != nil {
		return nil, d.options.InitialBackoff, err
	}
	dateStr := now.Format("2006-01-02")
	reqCtx, cancel := context.WithTimeout(ctx, d.options.RequestTimeout)
	defer cancel()
	wd, err := d.options.Fetcher.FetchWorkDay(reqCtx, dateStr)
	if err != nil {
		var rate *api.RateLimitError
		if errors.As(err, &rate) && rate.RetryAfter > 0 {
			return nil, rate.RetryAfter, err
		}
		return nil, d.options.InitialBackoff, err
	}
	plan, err := BuildDayPlan(now, wd, anchors, now)
	if err != nil {
		return nil, d.options.InitialBackoff, err
	}
	return &dayState{plan: plan, fetchedAt: now}, 0, nil
}

func (d *Daemon) sleep(ctx context.Context, duration time.Duration) bool {
	if duration < 0 {
		duration = 0
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (d *Daemon) logf(format string, args ...any) {
	if d.options.Logger != nil {
		d.options.Logger.Printf(format, args...)
	}
}

func nextBackoff(current, max time.Duration) time.Duration {
	if current <= 0 {
		current = defaultInitialBackoff
	}
	next := current * 2
	if max > 0 && next > max {
		return max
	}
	return next
}

func jitter(duration time.Duration) time.Duration {
	if duration <= time.Second {
		return duration
	}
	delta := duration / 10
	return duration - delta + time.Duration(rand.Int63n(int64(delta*2)))
}
