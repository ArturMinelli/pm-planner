package main

import (
	"context"
	"strings"
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

// TestAuth validates email/password with a fresh sign-in (ignores cached session).
// Returns an empty string on success, or a user-facing error message.
func (a *App) TestAuth(email, password string) string {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return "informe e-mail e senha"
	}
	if err := auth.VerifyCredentials(email, password); err != nil {
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
