package main

import (
	"github.com/go-pg/pg/orm"
	migrations "github.com/robinjoseph08/go-pg-migrations"
)

func init() {
	up := func(db orm.DB) error {
		_, err := db.Exec(`
			CREATE TABLE clusters
			(
				id                     TEXT PRIMARY KEY,
				environment            TEXT NOT NULL,
				server_url             TEXT NOT NULL,
				cluster_authority_data TEXT NOT NULL,
				deleted                BOOLEAN DEFAULT false,
				active                 BOOLEAN DEFAULT false,
				date_created           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				date_modified          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`)
		return err
	}

	down := func(db orm.DB) error {
		_, err := db.Exec("DROP TABLE clusters")
		return err
	}

	opts := migrations.MigrationOptions{}

	migrations.Register("20190610154448_craete_clusters_table", up, down, opts)
}
