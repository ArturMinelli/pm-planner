package server

import (
	"pm-cli/pkg/config"
	"pm-cli/pkg/message"
	"pm-cli/pkg/plan"
	"pm-cli/pkg/reminder"
)

// AuthResult is returned by TestAuth.
type AuthResult struct {
	OK    bool             `json:"ok"`
	Error *message.Message `json:"error,omitempty"`
}

// RecalculateRequest carries editable planner state from the frontend.
type RecalculateRequest struct {
	Date           string                  `json:"date"`
	BaseTargetSecs int64                   `json:"baseTargetSecs"`
	Balance        *plan.BalanceAdjustment `json:"balance,omitempty"`
	Journeys       []plan.Journey          `json:"journeys"`
	SolvedSlot     plan.SolvedSlot         `json:"solvedSlot"`
}

// ReminderStatus describes reminder configuration and runtime notification state.
type ReminderStatus struct {
	Settings                 config.Reminders `json:"settings"`
	AutostartEnabled         bool             `json:"autostartEnabled"`
	DaemonRunning            bool             `json:"daemonRunning"`
	NotificationAvailable    bool             `json:"notificationAvailable"`
	NotificationAuthorized   bool             `json:"notificationAuthorized"`
	NotificationStatusDetail string           `json:"notificationStatusDetail,omitempty"`
}

// ReminderPlanPayload is the reminder schedule for browser notification scheduling.
type ReminderPlanPayload struct {
	Schedule []reminder.ScheduledReminder `json:"schedule"`
}
