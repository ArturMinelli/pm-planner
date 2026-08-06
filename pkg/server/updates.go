package server

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
)

const updateCheckTimeout = 90 * time.Second

const keyDesktopOnlyUpdates = "common.desktopOnlyUpdates"

// DesktopOnlyUpdateStatus reports that self-update is unavailable outside the desktop shell.
func DesktopOnlyUpdateStatus() *update.Status {
	return &update.Status{
		Blockers: []message.Message{
			message.New(keyDesktopOnlyUpdates, nil),
		},
		UpdateAvailable: false,
	}
}

// CheckForUpdate reports the installed version and whether newer code exists.
func CheckForUpdate(ctx context.Context) (*update.Status, error) {
	binary, err := update.ResolvePMBinary()
	if err != nil {
		return &update.Status{Blockers: []message.Message{
			message.New(message.KeyUpdateBlockersPMNotFound, nil),
		}}, nil
	}

	if ctx == nil {
		ctx = context.Background()
	}
	reqCtx, cancel := context.WithTimeout(ctx, updateCheckTimeout)
	defer cancel()

	output, err := runPM(reqCtx, binary, "update", "--check", "--json")
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

// ConsumeUpdateResult returns the outcome of an update that ran while the app was closed.
func ConsumeUpdateResult() (*update.Result, error) {
	return update.ConsumeResult()
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
