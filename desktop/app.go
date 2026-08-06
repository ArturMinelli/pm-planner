package main

import (
	"context"

	"pm-cli/pkg/app"
	"pm-cli/pkg/config"
	"pm-cli/pkg/server"
	"pm-cli/pkg/update"
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
	// Updates are manual now; drop the daily schedule older installs registered.
	_ = update.RemoveScheduledAutoUpdate()
}

// GetConfig reads the YAML file on disk used by pm CLI (~/.config/pm/config.yaml).
func (a *App) GetConfig() (*config.File, error) {
	return server.GetConfig()
}

// SaveConfig writes YAML and refreshes viper so auth/session use the latest values.
func (a *App) SaveConfig(f *config.File) error {
	return server.SaveConfig(f)
}

// TestAuth validates login/password with a fresh sign-in (ignores cached session), then
// confirms the resulting session is actually accepted by the API.
func (a *App) TestAuth(login, password string) server.AuthResult {
	return server.TestAuth(a.context(), login, password)
}

// RecalculateRequest carries the editable planner state sent from the frontend.
type RecalculateRequest = server.RecalculateRequest

// RecalculateDay recomputes the solved slot and journey summaries from frontend-editable inputs.
func (a *App) RecalculateDay(req RecalculateRequest) (*app.PlannerSummary, error) {
	return server.RecalculateDay(req)
}

// LoadPlanner fetches the work day and builds suggested clock times (shared with CLI).
// On fetch failure it falls back to Settings defaults and sets LoadWarning.
func (a *App) LoadPlanner(date string) (*app.PlannerPayload, error) {
	return server.LoadPlanner(a.context(), date)
}

func (a *App) context() context.Context {
	if a.ctx == nil {
		return context.Background()
	}
	return a.ctx
}
