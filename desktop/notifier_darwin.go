//go:build darwin

package main

import (
	"context"
	"os/exec"
	"strings"
)

type systemNotifier struct{}

func (systemNotifier) Notify(ctx context.Context, title, body string) error {
	script := `display notification "` + appleScriptEscape(body) + `" with title "` + appleScriptEscape(title) + `"`
	return exec.CommandContext(ctx, "osascript", "-e", script).Run()
}

func appleScriptEscape(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, `"`, `\"`)
}
