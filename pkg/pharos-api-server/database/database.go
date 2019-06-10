package database

import (
	"crypto/tls"
	"fmt"

	"github.com/go-pg/pg"
	"github.com/lob/pharos/pkg/pharos-api-server/config"
)

// New initializes a new database struct.
func New(cfg config.Config) (*pg.DB, error) {
	opts := &pg.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.DatabaseHost, cfg.DatabasePort),
		Database: cfg.DatabaseName,
		Password: cfg.DatabasePassword,
		User:     cfg.DatabaseUser,
	}

	if cfg.DatabaseSSLMode {
		opts.TLSConfig = &tls.Config{ServerName: cfg.DatabaseHost}
	}

	db := pg.Connect(opts)

	// Ensure the database can connect.
	_, err := db.Exec("SELECT 1")
	if err != nil {
		return nil, err
	}
	return db, nil
}
