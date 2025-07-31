package cmd

import (
	"context"

	"github.com/grafana/seccomp-daemonset/cmd/healthcheck"
	"github.com/urfave/cli/v3"
)

func NewRoot() *cli.Command {
	return &cli.Command{
		Name:  "seccomp-daemonset",
		Usage: "Kubernetes operator for managing seccomp profiles",
		Commands: []*cli.Command{
			healthcheck.NewCmd(),
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return Run(ctx)
		},
	}
}

func Run(ctx context.Context) error {
	return nil
}
