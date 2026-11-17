package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"pm-cli/pkg/api"
	"pm-cli/pkg/config"
	pmserver "pm-cli/pkg/server"
	"pm-cli/pkg/server/httpapi"
)

func main() {
	// The dev API must never punch on the real PontoMais; use pm-fake-pontomais.
	api.BlockRealPunches()

	port := os.Getenv("PM_API_PORT")
	if port == "" {
		port = "3847"
	}

	if err := config.Init(""); err != nil {
		fmt.Fprintf(os.Stderr, "config init: %v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if api.FakeMode() {
		fmt.Printf("FAKE MODE: pontomais URL=%s\n", os.Getenv("PM_FAKE_PONTOMAIS_URL"))
	}

	go pmserver.PunchLog().Run(ctx)

	server := httpapi.New(port)
	fmt.Printf("PM dev API listening on http://127.0.0.1:%s\n", port)

	if err := server.ListenAndServe(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		os.Exit(1)
	}
}
