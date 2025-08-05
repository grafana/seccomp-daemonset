package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/grafana/seccomp-daemonset/cmd/healthcheck"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/semaphore"
)

func NewRoot() *cli.Command {
	return &cli.Command{
		Name:  "seccomp-daemonset",
		Usage: "Kubernetes operator for managing seccomp profiles",
		Commands: []*cli.Command{
			healthcheck.NewCmd(),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "source",
				Usage:     "Where should the seccomp profiles be read from?",
				Value:     "/profiles",
				TakesFile: true,
				Validator: isDir(true),
			},
			&cli.StringFlag{
				Name:      "destination",
				Usage:     "Where should the seccomp profiles be written to?",
				Value:     "/var/lib/kubelet/seccomp/profiles",
				TakesFile: true,
				Validator: isDir(false),
			},
			&cli.BoolWithInverseFlag{
				Name:  "mkdir",
				Usage: "Create the destination directory if it does not exist, including parents",
				Value: true,
			},
			&cli.BoolFlag{
				Name:  "one-shot",
				Usage: "Run the useful work, then exit without waiting on any HTTP requests. Still runs the server, but it won't block on it. Useful for testing.",
			},

			&cli.StringFlag{
				Name:  "listen-address",
				Usage: "The address to listen on for HTTP requests. If no address is specified, it will listen on all anycast and unicast interfaces.",
				Value: ":8080",
			},
			&cli.BoolFlag{
				Name:  "enable-default-collectors",
				Usage: "Should the Go and process collectors be enabled? This will add additional metrics and add CPU and memory requirements.",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			registry := prometheus.NewRegistry()
			if c.Bool("enable-default-collectors") {
				registry.MustRegister(collectors.NewGoCollector())
				registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
			}

			return Run(ctx, c.String("listen-address"), filepath.Clean(c.String("source")), filepath.Clean(c.String("destination")), c.Bool("mkdir"), registry, c.Bool("one-shot"))
		},
	}
}

type registry interface {
	prometheus.Registerer
	prometheus.Gatherer
}

func Run(ctx context.Context, listenAddr, source, destination string, mkdir bool, registry registry, oneShot bool) error {
	if mkdir {
		if err := os.MkdirAll(destination, 0750); err != nil {
			return cli.Exit(fmt.Errorf("failed to create destination directory %q: %w", destination, err), 1)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "health check requested, returning OK")
		_, _ = w.Write([]byte("Healthy"))
	})
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		ProcessStartTime: time.Now(),
	}))
	server := &http.Server{
		ReadTimeout:       time.Second * 3,
		WriteTimeout:      time.Second * 3,
		ReadHeaderTimeout: time.Second * 3,
		IdleTimeout:       time.Second * 30,
		Addr:              listenAddr,
		Handler:           mux,
	}
	defer func() {
		if err := server.Close(); err != nil {
			slog.Error("failed to close HTTP server", "err", err)
		}
	}()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- listenAndServe(server)
	}()
	go func() {
		<-ctx.Done()
		if err := server.Close(); err != nil {
			slog.Error("failed to close HTTP server on context cancellation", "err", err)
		}
	}()
	slog.Info("http server started", "address", listenAddr)

	if err := copyFiles(ctx, source, destination, registry); err != nil {
		if errors.Is(err, context.Canceled) {
			slog.Info("copying files interrupted by context cancellation, shutting down", "err", err)
			return nil
		}
		slog.Error("failed to copy files", "err", err)
		return fmt.Errorf("failed to copy files from %q to %q: %w", source, destination, err)
	}
	if oneShot {
		slog.Info("one-shot mode enabled, exiting after copying files")
		return nil
	}

	select {
	case <-ctx.Done():
		slog.Info("shutting down")
		return nil

	case err := <-serverErr:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			slog.Info("http server stopped gracefully")
			return nil
		}
		if errors.Is(err, context.Canceled) {
			slog.Info("http server stopped due to context cancellation", "err", err)
			return nil
		}
		slog.Error("http server stopped with error", "err", err)
		return fmt.Errorf("http server error: %w", err)
	}
}

func isDir(required bool) func(path string) error {
	return func(path string) error {
		if required && path == "" {
			return cli.Exit("path is required", 1)
		}

		stat, err := os.Stat(path)
		if os.IsNotExist(err) {
			if !required {
				return nil
			} else {
				return cli.Exit(fmt.Errorf("path %q does not exist", path), 1)
			}
		} else if err != nil {
			return cli.Exit(fmt.Errorf("error checking path %q: %w", path, err), 1)
		}

		if !stat.IsDir() {
			return cli.Exit(fmt.Errorf("path %q is not a directory", path), 1)
		}

		return nil
	}
}

func copyFiles(ctx context.Context, src, dst string, registry prometheus.Registerer) error {
	doneMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "seccomp_daemonset_complete",
		Help: "Indicates whether the copying process is complete",
	})
	registry.MustRegister(doneMetric)
	defer doneMetric.Set(1)

	countMetric := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "seccomp_daemonset_file_copy_count",
		Help: "Number of files copied",
	}, []string{"status", "path"})
	registry.MustRegister(countMetric)

	sema := semaphore.NewWeighted(64)
	wg := &sync.WaitGroup{}

	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return fmt.Errorf("context cancelled while walking directory: %w", ctx.Err())
		}
		if d.IsDir() {
			return nil // we don't care for directories
		}
		if shouldIgnore, why := shouldIgnore(path); shouldIgnore {
			slog.Debug("ignoring file", "path", path, "reason", why)
			countMetric.WithLabelValues("ignored", path).Inc()
			return nil
		}

		srcPath := filepath.Join(path)
		dstPath := filepath.Join(dst, strings.TrimPrefix(path, src))
		if srcPath == dstPath {
			slog.Warn("source and destination paths are the same, skipping", "path", srcPath)
			countMetric.WithLabelValues("failure", path).Inc()
			return nil
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := sema.Acquire(ctx, 1); err != nil {
				slog.Warn("failed to acquire semaphore", "err", err)
				return
			}
			defer sema.Release(1)

			slog.Info("copying file", "src", srcPath, "dst", dstPath)
			if err := duplicateIndividualFile(srcPath, dstPath); err != nil {
				slog.Error("failed to copy file", "src", srcPath, "dst", dstPath, "err", err)
				countMetric.WithLabelValues("failure", path).Inc()
			} else {
				slog.Info("file copied successfully", "src", srcPath, "dst", dstPath)
				countMetric.WithLabelValues("success", path).Inc()
			}
		}()
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk source directory %q: %w", src, err)
	}

	wait := make(chan struct{})
	go func() {
		wg.Wait()
		close(wait)
	}()

	select {
	case <-ctx.Done():
		slog.Info("context cancelled while copying files")
		return ctx.Err()
	case <-wait:
		slog.Info("all files copied successfully")
		return nil
	}
}

func shouldIgnore(path string) (bool, string) {
	if strings.HasPrefix(filepath.Base(path), "..") {
		return true, "file has .. prefix"
	}
	// If any element in the path also has a ".." prefix, we must ignore it, as this is probably a directory like "..data" or "..2025_08_04_06_45_56.476364577".
	for elem := range strings.SplitSeq(path, string(filepath.Separator)) {
		if strings.HasPrefix(elem, "..") {
			return true, fmt.Sprintf("file has element with .. prefix: %q", elem)
		}
	}

	return false, ""
}

func duplicateIndividualFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return fmt.Errorf("failed to create destination directory for %q: %w", dst, err)
	}

	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing file %q: %w", dst, err)
	}

	if err := linkFile(src, dst); err != nil {
		slog.Warn("failed to create hard link, falling back to copy", "src", src, "dst", dst, "err", err)
	}
	if err := copyFile(src, dst); err != nil {
		return fmt.Errorf("failed to copy file from %q to %q: %w", src, dst, err)
	}

	return nil
}

// linkFile attempts to create a hard link from src to dst.
func linkFile(src, dst string) error {
	if err := os.Link(src, dst); err != nil {
		return err
	}
	return nil
}

func copyFile(srcPath, dstPath string) error {
	// We don't want to leave half-finished files around, so we will temporarily ignore any and all context cancellations.
	// This is to ensure we don't cause kubelet any issues.

	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file %q: %w", srcPath, err)
	}
	defer func() {
		err := src.Close()
		if err != nil {
			slog.Warn("failed to close source file", "src", srcPath, "err", err)
		}
	}()

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %q: %w", dstPath, err)
	}
	defer func() {
		err := dst.Close()
		if err != nil {
			slog.Warn("failed to close destination file", "dst", dstPath, "err", err)
		}
	}()

	const BUF_SIZE = 16 * 1024
	buf := make([]byte, BUF_SIZE)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			if _, err := dst.Write(buf[:n]); err != nil {
				return fmt.Errorf("failed to write to destination file %q: %w", dstPath, err)
			}
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to read from source file %q: %w", srcPath, err)
		}
	}
}

func listenAndServe(server *http.Server) error {
	if strings.HasPrefix(server.Addr, "/") {
		// If the address starts with a slash, we assume it's a Unix domain socket.
		// We primarily support Unix sockets for tests.
		listener, err := net.Listen("unix", server.Addr)
		if err != nil {
			return fmt.Errorf("failed to listen on Unix socket %q: %w", server.Addr, err)
		}
		defer func() {
			if err := listener.Close(); err != nil {
				slog.Warn("failed to close Unix socket listener", "addr", server.Addr, "err", err)
			}
		}()
		server.Addr = "" // Clear the Addr field since we're using a listener
		return server.Serve(listener)
	} else {
		// Otherwise, it's probably a TCP address.
		return server.ListenAndServe()
	}
}
