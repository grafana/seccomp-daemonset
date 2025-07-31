package main

import (
	"context"
	"errors"
	"os"
	"os/signal"

	"github.com/grafana/seccomp-operator/cmd"
)

func main() {
	os.Exit(run())
}

func run() int {
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
