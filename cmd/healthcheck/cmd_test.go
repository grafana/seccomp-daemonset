package healthcheck_test

import (
	"testing"

	"github.com/grafana/seccomp-daemonset/cmd"
	"github.com/stretchr/testify/require"
)

func TestCommandConstruction(t *testing.T) {
	t.Parallel()

	err := cmd.NewRoot().Run(t.Context(), []string{"seccomp-daemonset", "healthcheck", "-h"})
	require.NoError(t, err, "should get help output without error")
}
