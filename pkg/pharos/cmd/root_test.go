package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Declare some constants to be used in testing functions used in various commands.
const (
	cliConfig       = "../testdata/pharosConfig"
	config          = "../testdata/config"
	malformedConfig = "../testdata/malformed"
	emptyConfig     = "../testdata/empty"
)

func TestArgID(t *testing.T) {
	t.Run("succeeds when we are given exactly one argument", func(tt *testing.T) {
		err := argID([]string{"1"})
		require.NoError(tt, err)
	})

	t.Run("errors if too few arguments are passed in", func(tt *testing.T) {
		err := argID([]string{})
		require.Error(tt, err)
	})

	t.Run("errors if too many arguments are passed in", func(tt *testing.T) {
		err := argID([]string{"1", "2"})
		require.Error(tt, err)
	})
}
