package cmd

import (
	"os"
	"testing"

	"github.com/lob/pharos/internal/test"
	configpkg "github.com/lob/pharos/pkg/pharos/config"
	"github.com/stretchr/testify/assert"
)

func TestRunSetup(t *testing.T) {
	t.Run("successfully initializes the pharos config file", func(tt *testing.T) {
		// Create temporary test config file and defer cleanup.
		configFile := test.CopyTestFile(tt, "../testdata", "setup", cliConfig)
		defer os.Remove(configFile)

		// Setup file.
		err := runSetup(configFile, "egg", "hello", "")
		assert.NoError(tt, err)

		// Check that file setup was successful.
		c, err := configpkg.New(configFile)
		assert.NoError(tt, err)
		err = c.Load()
		assert.NoError(tt, err)

		assert.Equal(tt, "egg", c.BaseURL)
		assert.Equal(tt, "hello", c.AWSProfile)
		assert.Equal(tt, "", c.AssumeRoleARN)
	})
}
