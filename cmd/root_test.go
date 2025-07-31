package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandConstruction(t *testing.T) {
	t.Parallel()

	err := NewRoot().Run(t.Context(), []string{"seccomp-operator", "-h"})
	require.NoError(t, err, "should get help output without error")
}
