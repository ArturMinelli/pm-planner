//go:build linux

package main

import (
	"context"
	"os/exec"
)

type systemNotifier struct{}

func (systemNotifier) Notify(ctx context.Context, title, body string) error {
	return exec.CommandContext(ctx, "notify-send", "-a", "PM Planner", title, body).Run()
}
