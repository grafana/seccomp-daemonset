package healthcheck

import (
	"context"

	"github.com/urfave/cli/v3"
)

func NewCmd() *cli.Command {
	return &cli.Command{
		Name:  "healthcheck",
		Usage: "Check if the server running here is healthy",
		Action: func(ctx context.Context, c *cli.Command) error {
			return Run(ctx)
		},
	}
}

func Run(ctx context.Context) error {
	return nil
}
