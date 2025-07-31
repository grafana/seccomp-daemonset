package healthcheck

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/urfave/cli/v3"
)

func NewCmd() *cli.Command {
	return &cli.Command{
		Name:  "healthcheck",
		Usage: "Check if the server running here is healthy",
		Flags: []cli.Flag{
			&cli.DurationFlag{
				Name:  "timeout",
				Usage: "Timeout for the health check request",
				Value: 5 * time.Second,
			},
			&cli.StringFlag{
				Name:  "listen-address",
				Usage: "The address the server is listening on",
				Value: "http://:8080",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return Run(ctx, c.String("listen-address"), c.Duration("timeout"))
		},
	}
}

func Run(ctx context.Context, listenAddr string, timeout time.Duration) error {
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listenAddr+"/healthz", nil)
	if err != nil {
		return cli.Exit("failed to create health check request: "+err.Error(), 1)
	}
	req.Header.Add("Accept", "text/plain")
	req.Header.Add("User-Agent", "seccomp-daemonset-healthcheck/Grafana Labs")
	resp, err := client.Do(req)
	if err != nil {
		return cli.Exit("failed to perform health check: "+err.Error(), 1)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			slog.Warn("failed to close response body", "err", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return cli.Exit("health check failed with status code: "+resp.Status, 1)
	}

	slog.Info("health check OK", "status", resp.Status)
	return nil
}
