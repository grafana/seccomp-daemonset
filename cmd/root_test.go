package cmd

import (
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCommandConstruction(t *testing.T) {
	t.Parallel()

	err := NewRoot().Run(t.Context(), []string{"seccomp-daemonset", "-h"})
	require.NoError(t, err, "should get help output without error")
}

func TestServerStarts(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	socket := filepath.Join(tempDir, "socket")
	src := filepath.Join(tempDir, "source")
	require.NoError(t, os.MkdirAll(src, 0755), "should create source directory")
	dst := filepath.Join(tempDir, "destination")

	go func() {
		err := NewRoot().Run(t.Context(), []string{"seccomp-daemonset", "--listen-address", socket, "--source", src, "--destination", dst})
		require.NoError(t, err, "should start server without error")
	}()
	client := unixClient(socket)
	t.Cleanup(client.CloseIdleConnections)
	require.Eventually(t, func() bool {
		resp, err := client.Get("http://unix/healthz")
		if err != nil {
			return false
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		return resp.StatusCode == http.StatusOK
	}, 3*time.Second, 100*time.Millisecond, "HTTP server should start within 3 seconds")
}

func TestServerServesMetrics(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	socket := filepath.Join(tempDir, "socket")
	src := filepath.Join(tempDir, "source")
	require.NoError(t, os.MkdirAll(src, 0755), "should create source directory")
	dst := filepath.Join(tempDir, "destination")

	go func() {
		err := NewRoot().Run(t.Context(), []string{"seccomp-daemonset", "--listen-address", socket, "--source", src, "--destination", dst})
		require.NoError(t, err, "should start server without error")
	}()
	client := unixClient(socket)
	t.Cleanup(client.CloseIdleConnections)
	require.Eventually(t, func() bool {
		resp, err := client.Get("http://unix/metrics")
		if err != nil {
			return false
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		return resp.StatusCode == http.StatusOK
	}, 3*time.Second, 100*time.Millisecond, "HTTP server should start within 3 seconds")
}

func TestCopiesFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	socket := filepath.Join(tempDir, "socket")
	src := filepath.Join(tempDir, "source")
	require.NoError(t, os.MkdirAll(src, 0755), "should create source directory")
	dst := filepath.Join(tempDir, "destination")

	require.NoError(t, os.MkdirAll(filepath.Join(src, "..2025_08_04_06_32_35.3325235941"), 0755), "creating ..2025_08_04_06_32_35.3325235941 dir in src")
	require.NoError(t, os.Symlink(filepath.Join(src, "..2025_08_04_06_32_35.3325235941"), filepath.Join(src, "..data")), "symlinking ..data to ..2025_08_04_06_32_35.3325235941 in src")
	require.NoError(t, os.WriteFile(filepath.Join(src, "..2025_08_04_06_32_35.3325235941", "profile.json"), []byte("test"), 0644), "writing profile.json in ..2025_08_04_06_32_35.3325235941 in src")
	require.NoError(t, os.Symlink(filepath.Join(src, "..2025_08_04_06_32_35.3325235941", "profile.json"), filepath.Join(src, "profile.json")), "symlinking profile.json into src")

	err := NewRoot().Run(t.Context(), []string{"seccomp-daemonset", "--listen-address", socket, "--source", src, "--destination", dst, "--one-shot"})
	require.NoError(t, err, "should start server without error")

	require.DirExists(t, dst, "destination directory should exist")
	require.FileExists(t, filepath.Join(dst, "profile.json"), "profile.json should be copied to destination")
	require.NoDirExists(t, filepath.Join(dst, "..2025_08_04_06_32_35.3325235941"), "should not copy .. dirs to destination")
	require.NoDirExists(t, filepath.Join(dst, "..data"), "should not copy ..dirs to destination")
}

func unixClient(socket string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _network, _addr string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", socket)
			},
		},
	}
}
