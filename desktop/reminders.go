package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"pm-cli/pkg/config"
	"pm-cli/pkg/reminder"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type ReminderStatus struct {
	Settings                 config.Reminders `json:"settings"`
	AutostartEnabled         bool             `json:"autostartEnabled"`
	DaemonRunning            bool             `json:"daemonRunning"`
	NotificationAvailable    bool             `json:"notificationAvailable"`
	NotificationAuthorized   bool             `json:"notificationAuthorized"`
	NotificationStatusDetail string           `json:"notificationStatusDetail,omitempty"`
}

type NotificationPermissionStatus struct {
	Available  bool   `json:"available"`
	Authorized bool   `json:"authorized"`
	Detail     string `json:"detail,omitempty"`
}

type desktopAlerter struct {
	native nativeNotifier
}

type nativeNotifier interface {
	Notify(ctx context.Context, title, body string) error
}

func (a desktopAlerter) SendReminder(ctx context.Context, event reminder.ScheduledReminder, settings config.Reminders) error {
	title := fmt.Sprintf("%s em %d min", event.SlotLabel, event.LeadMinutes)
	body := fmt.Sprintf("Horário recomendado: %s", event.SlotTime.Format("15:04"))

	var errs []error
	sent := false
	if config.ReminderNativeEnabled(settings) && a.native != nil {
		if err := a.native.Notify(ctx, title, body); err != nil {
			errs = append(errs, err)
		} else {
			sent = true
		}
	}
	if sent {
		return nil
	}
	if len(errs) == 0 {
		return fmt.Errorf("no notification channel enabled")
	}
	return errors.Join(errs...)
}

func (a *App) GetReminderStatus() (*ReminderStatus, error) {
	f, err := config.Read("")
	if err != nil {
		return nil, err
	}
	settings, err := config.ResolveReminders(f)
	if err != nil {
		return nil, err
	}
	autostartEnabled, _ := autostartIsEnabled()
	running, _ := daemonRunning()
	permission := a.notificationPermissionStatus()
	return &ReminderStatus{
		Settings:                 settings,
		AutostartEnabled:         autostartEnabled,
		DaemonRunning:            running,
		NotificationAvailable:    permission.Available,
		NotificationAuthorized:   permission.Authorized,
		NotificationStatusDetail: permission.Detail,
	}, nil
}

func (a *App) SaveReminderSettings(settings config.Reminders) error {
	normalized, err := config.ResolveReminders(&config.File{Reminders: &settings})
	if err != nil {
		return err
	}
	f, err := config.Read("")
	if err != nil {
		return err
	}
	f.Reminders = &normalized
	if err := config.Save("", f); err != nil {
		return err
	}

	executable, err := os.Executable()
	if err != nil {
		return err
	}
	if normalized.Enabled && normalized.Autostart {
		if err := autostartEnable(executable); err != nil {
			return err
		}
	} else {
		if err := autostartDisable(); err != nil {
			return err
		}
	}

	if normalized.Enabled {
		return ensureDaemonRunning(executable)
	}
	return stopDaemon()
}

func (a *App) RequestNotificationPermission() (*NotificationPermissionStatus, error) {
	if a.ctx == nil {
		return &NotificationPermissionStatus{Available: false, Authorized: false, Detail: "runtime indisponível"}, nil
	}
	if err := wailsruntime.InitializeNotifications(a.ctx); err != nil {
		return nil, err
	}
	available := wailsruntime.IsNotificationAvailable(a.ctx)
	if !available {
		return &NotificationPermissionStatus{Available: false, Authorized: false, Detail: "notificações não disponíveis"}, nil
	}
	authorized, err := wailsruntime.RequestNotificationAuthorization(a.ctx)
	if err != nil {
		return nil, err
	}
	return &NotificationPermissionStatus{Available: available, Authorized: authorized}, nil
}

func (a *App) SendTestReminder() error {
	f, err := config.Read("")
	if err != nil {
		return err
	}
	settings, err := config.ResolveReminders(f)
	if err != nil {
		return err
	}
	now := time.Now()
	event := reminder.ScheduledReminder{
		ID:          reminder.ReminderID(now.Format("2006-01-02"), reminder.SlotOut2, 5),
		Date:        now.Format("2006-01-02"),
		SlotKey:     reminder.SlotOut2,
		SlotLabel:   "Saída 2",
		SlotTime:    now.Add(5 * time.Minute),
		LeadMinutes: 5,
		FireAt:      now,
	}
	return desktopAlerter{
		native: systemNotifier{},
	}.SendReminder(context.Background(), event, settings)
}

func (a *App) notificationPermissionStatus() NotificationPermissionStatus {
	if a.ctx == nil {
		return NotificationPermissionStatus{Available: false, Authorized: false, Detail: "runtime indisponível"}
	}
	if err := wailsruntime.InitializeNotifications(a.ctx); err != nil {
		return NotificationPermissionStatus{Available: false, Authorized: false, Detail: err.Error()}
	}
	available := wailsruntime.IsNotificationAvailable(a.ctx)
	authorized := false
	if available {
		if ok, err := wailsruntime.CheckNotificationAuthorization(a.ctx); err == nil {
			authorized = ok
		}
	}
	return NotificationPermissionStatus{Available: available, Authorized: authorized}
}

func writeDaemonPID() error {
	path, err := reminder.DefaultPIDPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0600)
}

func removeDaemonPID() {
	path, err := reminder.DefaultPIDPath()
	if err == nil {
		_ = os.Remove(path)
	}
}

func readDaemonPID() (int, error) {
	path, err := reminder.DefaultPIDPath()
	if err != nil {
		return 0, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || pid <= 0 {
		return 0, fmt.Errorf("invalid daemon pid")
	}
	return pid, nil
}

func daemonRunning() (bool, error) {
	pid, err := readDaemonPID()
	if err != nil {
		return false, nil
	}
	return processRunning(pid), nil
}

func ensureDaemonRunning(executable string) error {
	if running, _ := daemonRunning(); running {
		return nil
	}
	cmd := exec.Command(executable, "--daemon")
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}

func stopDaemon() error {
	pid, err := readDaemonPID()
	if err != nil {
		return nil
	}
	if err := terminateProcess(pid); err != nil {
		return err
	}
	removeDaemonPID()
	return nil
}
