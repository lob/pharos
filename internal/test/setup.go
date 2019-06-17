package test

import (
	"testing"

	"github.com/go-pg/pg"
	"github.com/stretchr/testify/require"
)

// TruncateTables truncates all of the tables in the database to be able to
// start from scratch for a test.
func TruncateTables(t *testing.T, db *pg.DB) {
	t.Helper()

	_, err := db.Exec(`
		TRUNCATE clusters CASCADE;
	`)
	require.NoError(t, err)
}
