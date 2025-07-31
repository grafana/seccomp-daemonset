package cmd

import (
	"github.com/grafana/seccomp-operator/cmd/daemon"
	"github.com/grafana/seccomp-operator/cmd/healthcheck"
	"github.com/grafana/seccomp-operator/cmd/operator"
	"github.com/urfave/cli/v3"
)

func NewRoot() *cli.Command {
	return &cli.Command{
		Name:  "seccomp-operator",
		Usage: "Kubernetes operator for managing seccomp profiles",
		Commands: []*cli.Command{
			daemon.NewCmd(),
			healthcheck.NewCmd(),
			operator.NewCmd(),
		},
	}
}
