package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func touchFile(t *testing.T, path string) {
	t.Helper()

	file, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, file.Close())
}
