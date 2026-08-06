package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"pm-cli/pkg/message"
	"pm-cli/pkg/update"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	updateCheckTimeout = 90 * time.Second
	// quitDelay lets the frontend render "closing to update" before the window
	// disappears.
	quitDelay = 1200 * time.Millisecond
)

// CheckForUpdate reports the installed version and whether newer code exists.
// Everything the user could act on comes back as a blocker inside the status,
// so the Settings card can always render something meaningful.
func (a *App) CheckForUpdate() (*update.Status, error) {
	binary, err := update.ResolvePMBinary()
	if err != nil {
		return &update.Status{Blockers: []message.Message{
			message.New(message.KeyUpdateBlockersPMNotFound, nil),
		}}, nil
	}

	ctx := a.context()
	ctx, cancel := context.WithTimeout(ctx, updateCheckTimeout)
	defer cancel()

	output, err := runPM(ctx, binary, "update", "--check", "--json")
	if err != nil {
		return &update.Status{Blockers: []message.Message{
			message.New(message.KeyUpdateBlockersCheckFailed, map[string]string{
				"detail": err.Error(),
			}),
		}}, nil
	}

	status := &update.Status{}
	if err := json.Unmarshal(output, status); err != nil {
		return nil, fmt.Errorf("unexpected pm update --check response: %w", err)
	}
	return status, nil
}

// StartUpdate hands the update to a detached `pm update` process and closes the
// app. The update script terminates pm-desktop before rebuilding it, so the
// updater cannot be a child of this process. It reopens the app when it is done.
func (a *App) StartUpdate() error {
	binary, err := update.ResolvePMBinary()
	if err != nil {
		return err
	}
	root, err := update.ResolveInstallRoot()
	if err != nil {
		return err
	}
	if err := update.SpawnDetached(root, binary, "update", "--apply", "--relaunch"); err != nil {
		return err
	}

	go func() {
		time.Sleep(quitDelay)
		wailsruntime.Quit(a.context())
	}()
	return nil
}

// ConsumeUpdateResult returns the outcome of an update that ran while the app
// was closed, exactly once. It returns nil when no update has run since.
func (a *App) ConsumeUpdateResult() (*update.Result, error) {
	return update.ConsumeResult()
}

func (a *App) context() context.Context {
	if a.ctx == nil {
		return context.Background()
	}
	return a.ctx
}

func runPM(ctx context.Context, binary string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, binary, args...)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		if message := stderr.String(); message != "" {
			return nil, fmt.Errorf("%s", firstLine(message))
		}
		return nil, err
	}
	return stdout.Bytes(), nil
}

func firstLine(value string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(value), "\n")
	return line
}
