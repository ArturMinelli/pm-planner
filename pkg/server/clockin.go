package server

import (
	"context"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/app"
)

// ClockIn registers a punch using the shared client (fake mode aware).
func ClockIn(ctx context.Context, req app.ClockInRequest) app.ClockInResult {
	if ctx == nil {
		ctx = context.Background()
	}
	reqCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	return app.ClockIn(reqCtx, api.New(), req)
}
