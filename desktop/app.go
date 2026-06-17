package main

import (
	"context"
	"strings"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/app"
	"pm-cli/pkg/auth"
	"pm-cli/pkg/config"
	"pm-cli/pkg/location"
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
	_, _ = auth.GetAuthHeaders()
}

// GetConfig reads the YAML file on disk used by pm CLI (~/.config/pm/config.yaml).
func (a *App) GetConfig() (*config.File, error) {
	return config.Read("")
}

// SaveConfig writes YAML and refreshes viper so auth/session use the latest values.
func (a *App) SaveConfig(f *config.File) error {
	return config.Save("", f)
}

// TestAuth validates login/password with a fresh sign-in (ignores cached session).
// Returns an empty string on success, or a user-facing error message.
func (a *App) TestAuth(login, password string) string {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return "informe login (e-mail ou CPF) e senha"
	}
	if err := auth.VerifyCredentials(login, password); err != nil {
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

// GetClockInStatus returns the most recent cached punch from PontoMais.
func (a *App) GetClockInStatus() (*api.ClockInStatus, error) {
	if a.ctx == nil {
		a.ctx = context.Background()
	}
	cl := api.New()
	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()
	return cl.FetchClockInStatus(ctx)
}

// RegisterTimeCard records a punch with the given geolocation.
func (a *App) RegisterTimeCard(latitude, longitude, accuracy float64, address string) (*api.RegisterTimeCardResult, error) {
	if a.ctx == nil {
		a.ctx = context.Background()
	}
	cl := api.New()
	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()
	return cl.RegisterTimeCard(ctx, api.TimeCardLocation{
		Latitude:  latitude,
		Longitude: longitude,
		Address:   strings.TrimSpace(address),
		Accuracy:  accuracy,
	})
}

// GetDeviceLocation returns coordinates from the host OS (GeoClue on Linux).
func (a *App) GetDeviceLocation() (*location.DeviceLocation, error) {
	if a.ctx == nil {
		a.ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(a.ctx, 20*time.Second)
	defer cancel()
	return location.GetDeviceLocation(ctx)
}
