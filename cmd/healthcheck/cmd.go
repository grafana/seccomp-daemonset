package healthcheck

import "github.com/urfave/cli/v3"

func NewCmd() *cli.Command {
	return &cli.Command{
		Name:  "healthcheck",
		Usage: "Check if the server running here is healthy",
	}
}
