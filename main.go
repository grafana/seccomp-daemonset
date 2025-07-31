package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"

	"github.com/grafana/seccomp-daemonset/cmd"
)

func main() {
	os.Exit(run())
}

func run() int {
	initLogger()

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	if err := cmd.NewRoot().Run(ctx, os.Args); err != nil {
		if errors.Is(err, context.Canceled) {
			// This is a graceful shutdown, so we can ignore the error.
			return 0
		}
		return 1
	}

	return 0
}

func initLogger() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
