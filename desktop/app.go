package main

import (
	"context"
	"time"

	"pm-cli/pkg/app"
	"pm-cli/pkg/api"
	"pm-cli/pkg/auth"
	"pm-cli/pkg/config"
)

// App is the Wails bind target; methods are exposed to the React frontend.
type App struct {
	ctx context.Context
}

// NewApp constructs the desktop app context.
func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	_ = config.Init("")
}

// GetConfig reads the YAML file on disk used by pm CLI (~/.config/pm/config.yaml).
func (a *App) GetConfig() (*config.File, error) {
	return config.Read("")
}

// SaveConfig writes YAML and refreshes viper so auth/session use the latest values.
func (a *App) SaveConfig(f *config.File) error {
	return config.Save("", f)
}

// PingAuth validates credentials/session without returning secrets (empty OK message on success).
func (a *App) PingAuth() string {
	if _, err := auth.GetAuthHeaders(); err != nil {
		return err.Error()
	}
	return ""
}

// LoadPlanner fetches the work day and builds suggested clock times (shared with CLI).
func (a *App) LoadPlanner(date string) (*app.PlannerPayload, error) {
	if a.ctx == nil {
		a.ctx = context.Background()
	}
	cl := api.New()
	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()
	return app.FetchPlannerPayload(ctx, cl, date)
}

// RecalculatePlanner recomputes saída 2 and segment summaries after edits.
func (a *App) RecalculatePlanner(date string, targetSecs int64, in1 string, out1 string, in2 string) (*app.PlannerSummary, error) {
	return app.RecalculatePlanner(date, targetSecs, in1, out1, in2)
}
