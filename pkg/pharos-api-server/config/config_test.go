package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	originalAdminAccessRoles := os.Getenv("ADMIN_ACCESS_ROLES")
	originalReadAccessRoles := os.Getenv("READ_ACCESS_ROLES")
	originalRobotAccessRoles := os.Getenv("ROBOT_ACCESS_ROLES")
	defer func() {
		err := os.Setenv("ADMIN_ACCESS_ROLES", originalAdminAccessRoles)
		require.Nil(t, err, "unexpected error restoring original ADMIN_ACCESS_ROLES")
		err = os.Setenv("READ_ACCESS_ROLES", originalReadAccessRoles)
		require.Nil(t, err, "unexpected error restoring original READ_ACCESS_ROLES")
		err = os.Setenv("ROBOT_ACCESS_ROLES", originalRobotAccessRoles)
		require.Nil(t, err, "unexpected error restoring original ROBOT_ACCESS_ROLES")
	}()

	err := os.Setenv("ADMIN_ACCESS_ROLES", "admin")
	require.Nil(t, err, "unexpected error setting test env value for ADMIN_ACCESS_ROLES")
	err = os.Setenv("READ_ACCESS_ROLES", "read1,read2")
	require.Nil(t, err, "unexpected error setting test env value for READ_ACCESS_ROLES")
	err = os.Setenv("ROBOT_ACCESS_ROLES", "robot")
	require.Nil(t, err, "unexpected error setting test env value for ROBOT_ACCESS_ROLES")

	cfg := New()
	assert.Equal(t, 7654, cfg.Port)
	assert.NotNil(t, cfg, "returned config shouldn't be nil")
	assert.Equal(t, []string{"admin"}, cfg.Permissions.Admin)
	assert.Equal(t, []string{"read1", "read2", "admin"}, cfg.Permissions.Read)
	assert.Equal(t, []string{"robot", "admin"}, cfg.Permissions.Robot)
}
