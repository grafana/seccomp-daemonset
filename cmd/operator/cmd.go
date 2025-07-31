package operator

import "github.com/urfave/cli/v3"

func NewCmd() *cli.Command {
	return &cli.Command{
		Name:  "operator",
		Usage: "Run the operator server",
	}
}
