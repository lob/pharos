package config

import (
	"os"
)

// Config contains the environment specific configuration values needed by the
// application.
type Config struct {
	DatabaseHost     string
	DatabasePort     int
	DatabaseName     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseSSLMode  bool
	Environment      string
	Hostname         string
	Port             int
	SentryDSN        string
	StatsdHost       string
	StatsdPort       int
}

const env = "ENVIRONMENT"

// New returns an instance of Config
func New() Config {
	cfg := Config{
		Hostname:         os.Getenv("HOSTNAME"),
		Port:             7654,
		Environment:      os.Getenv(env),
		SentryDSN:        os.Getenv("SENTRY_DSN"),
		StatsdHost:       "127.0.0.1",
		StatsdPort:       8125,
		DatabaseHost:     os.Getenv("DATABASE_HOST"),
		DatabasePort:     5432,
		DatabaseName:     os.Getenv("DATABASE_NAME"),
		DatabaseUser:     os.Getenv("DATABASE_USER"),
		DatabasePassword: os.Getenv("DATABASE_PASSWORD"),
		DatabaseSSLMode:  true,
	}

	switch os.Getenv(env) {
	case "development", "":
		cfg.DatabaseHost = "127.0.0.1"
		cfg.DatabaseName = "pharos"
		cfg.DatabaseUser = "pharos_admin"
		cfg.DatabaseSSLMode = false
	case "test":
		cfg.DatabaseHost = "127.0.0.1"
		cfg.DatabaseName = "pharos_test"
		cfg.DatabaseUser = "pharos_admin"
		cfg.DatabaseSSLMode = false
	}

	return cfg
}
