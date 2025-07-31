package daemon

import "github.com/urfave/cli/v3"

func NewCmd() *cli.Command {
	return &cli.Command{
		Name:  "daemon",
		Usage: "Run the per-node daemon",
	}
}
