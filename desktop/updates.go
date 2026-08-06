package main

import (
	"time"

	"pm-cli/pkg/server"
	"pm-cli/pkg/update"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	// quitDelay lets the frontend render "closing to update" before the window
	// disappears.
	quitDelay = 1200 * time.Millisecond
)

// CheckForUpdate reports the installed version and whether newer code exists.
// Everything the user could act on comes back as a blocker inside the status,
// so the Settings card can always render something meaningful.
func (a *App) CheckForUpdate() (*update.Status, error) {
	return server.CheckForUpdate(a.context())
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
	return server.ConsumeUpdateResult()
}
