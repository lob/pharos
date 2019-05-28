package config

import (
	"os"
)

// Config contains the environment specific configuration values needed by the
// application.
type Config struct {
	Environment string
	Hostname    string
	Port        int
	SentryDSN   string
	StatsdHost  string
	StatsdPort  int
}

// New returns an instance of Config
func New() Config {
	cfg := Config{
		Hostname:    os.Getenv("HOSTNAME"),
		Port:        7654,
		Environment: os.Getenv("ENVIRONMENT"),
		SentryDSN:   os.Getenv("SENTRY_DSN"),
		StatsdHost:  "127.0.0.1",
		StatsdPort:  8125,
	}

	return cfg
}
