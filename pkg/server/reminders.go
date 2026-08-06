package server

import (
	"context"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/config"
	"pm-cli/pkg/reminder"
)

// ReadReminderSettings loads reminder settings from config.
func ReadReminderSettings() (config.Reminders, error) {
	file, err := config.Read("")
	if err != nil {
		return config.Reminders{}, err
	}
	return config.ResolveReminders(file)
}

// SaveReminderSettingsToConfig normalizes and persists reminder settings only.
func SaveReminderSettingsToConfig(settings config.Reminders) error {
	normalized, err := config.ResolveReminders(&config.File{Reminders: &settings})
	if err != nil {
		return err
	}
	file, err := config.Read("")
	if err != nil {
		return err
	}
	file.Reminders = &normalized
	return config.Save("", file)
}

// BrowserReminderStatus returns reminder status for the HTTP / dev-browser path.
// Notification permission is determined client-side via the Web Notifications API.
func BrowserReminderStatus() (*ReminderStatus, error) {
	settings, err := ReadReminderSettings()
	if err != nil {
		return nil, err
	}
	return &ReminderStatus{
		Settings:                 settings,
		AutostartEnabled:         false,
		DaemonRunning:            false,
		NotificationAvailable:    true,
		NotificationAuthorized:   false,
		NotificationStatusDetail: "",
	}, nil
}

// BuildReminderPlan fetches the work day and builds the reminder schedule for a date.
func BuildReminderPlan(ctx context.Context, dateStr string) (*ReminderPlanPayload, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	settings, err := ReadReminderSettings()
	if err != nil {
		return nil, err
	}

	file, err := config.Read("")
	if err != nil {
		return nil, err
	}
	anchors, err := config.ResolvePlannerAnchors(file)
	if err != nil {
		return nil, err
	}

	loc := time.Now().Location()
	date, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return nil, err
	}

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	client := api.New()
	workDay, err := client.FetchWorkDay(reqCtx, dateStr)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	dayPlan, err := reminder.BuildDayPlan(date, workDay, anchors, now)
	if err != nil {
		return nil, err
	}

	schedule := reminder.BuildSchedule(dayPlan, settings)
	return &ReminderPlanPayload{Schedule: schedule}, nil
}
