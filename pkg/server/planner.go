package server

import (
	"context"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/app"
	"pm-cli/pkg/message"
)

// LoadPlanner fetches the work day and builds suggested clock times.
// On fetch failure it falls back to settings defaults and sets LoadWarning.
func LoadPlanner(ctx context.Context, date string) (*app.PlannerPayload, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	client := api.New()
	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	payload, err := app.FetchPlannerPayload(reqCtx, client, date)
	if err == nil {
		return payload, nil
	}

	payload, defaultsErr := app.BuildDefaultsPlannerPayload(date)
	if defaultsErr != nil {
		return nil, err
	}
	payload.LoadWarning = message.Ptr(message.KeyErrorsPlannerLoadFallback, nil)
	return payload, nil
}

// RecalculateDay recomputes the solved slot and journey summaries from editable inputs.
func RecalculateDay(req RecalculateRequest) (*app.PlannerSummary, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, err
	}
	return app.RecalculatePlanner(date, req.BaseTargetSecs, req.Balance, req.Journeys, req.SolvedSlot)
}
