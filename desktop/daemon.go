package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"pm-cli/pkg/api"
	"pm-cli/pkg/config"
	"pm-cli/pkg/reminder"
)

func runDaemon() error {
	if err := config.Init(""); err != nil {
		return err
	}
	f, err := config.Read("")
	if err != nil {
		return err
	}
	settings, err := config.ResolveReminders(f)
	if err != nil {
		return err
	}
	if !settings.Enabled {
		return nil
	}
	if running, _ := daemonRunning(); running {
		return nil
	}
	if err := writeDaemonPID(); err != nil {
		return err
	}
	defer removeDaemonPID()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	executable, _ := os.Executable()
	statePath, err := reminder.DefaultStatePath()
	if err != nil {
		return err
	}
	daemon := reminder.NewDaemon(reminder.DaemonOptions{
		Fetcher: api.New(),
		Alerter: desktopAlerter{
			executable: executable,
			native:     systemNotifier{},
		},
		Store:  reminder.NewFileStore(statePath),
		Logger: log.New(os.Stderr, "pm-reminders: ", log.LstdFlags),
	})
	err = daemon.Run(ctx)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}
