package server

import (
	"context"
	"strings"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/auth"
	"pm-cli/pkg/config"
	"pm-cli/pkg/message"
)

// GetConfig reads the YAML config file (~/.config/pm/config.yaml).
func GetConfig() (*config.File, error) {
	return config.Read("")
}

// SaveConfig writes YAML and refreshes viper so auth/session use the latest values.
func SaveConfig(file *config.File) error {
	return config.Save("", file)
}

// TestAuth validates login/password with a fresh sign-in, then confirms API access.
func TestAuth(ctx context.Context, login, password string) AuthResult {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return AuthResult{
			Error: message.Ptr(message.KeyErrorsAuthMissingCredentials, nil),
		}
	}
	if err := auth.VerifyCredentials(login, password); err != nil {
		classified := message.FromError(err)
		return AuthResult{Error: &classified}
	}

	if ctx == nil {
		ctx = context.Background()
	}
	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := api.New().VerifyAccess(reqCtx); err != nil {
		classified := message.FromError(err)
		return AuthResult{Error: &classified}
	}
	return AuthResult{OK: true}
}
