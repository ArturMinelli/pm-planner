package main

import (
	"context"
	"strings"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/app"
	"pm-cli/pkg/auth"
	"pm-cli/pkg/config"
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
	return config.Read("")
}

// SaveConfig writes YAML and refreshes viper so auth/session use the latest values.
func (a *App) SaveConfig(f *config.File) error {
	return config.Save("", f)
}

// TestAuth validates login/password with a fresh sign-in (ignores cached session), then
// confirms the resulting session is actually accepted by the API. A successful sign-in
// alone is not sufficient: PontoMais can accept credentials at /auth/sign_in while still
// rejecting the very next authenticated request if the session headers it issued aren't
// the ones the app ends up using (see pkg/api.Client.DoAuthenticated).
// Returns an empty string on success, or a user-facing error message.
func (a *App) TestAuth(login, password string) string {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return "informe login (e-mail ou CPF) e senha"
	}
	if err := auth.VerifyCredentials(login, password); err != nil {
		return err.Error()
	}

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := api.New().VerifyAccess(reqCtx); err != nil {
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
