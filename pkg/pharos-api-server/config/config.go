package config

import (
	"os"
	"strings"
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
	Permissions      *Permissions
	SentryDSN        string
	StatsdHost       string
	StatsdPort       int
}

// Permissions contains lists of AWS IAM ARNs that are to be associated with
// each of the 3 valid permission groups.
type Permissions struct {
	Admin []string
	Read  []string
	Write []string
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
		Permissions:      &Permissions{},
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

	// Load admin IAM roles
	if adminRoles := os.Getenv("ADMIN_ACCESS_ROLES"); adminRoles != "" {
		cfg.Permissions.Admin = strings.Split(adminRoles, ",")
	}

	// Load read IAM roles
	if readRoles := os.Getenv("READ_ACCESS_ROLES"); readRoles != "" {
		// As admins are allowed to perform any action we append their roles to the
		// Read list of ARNs to prevent having to add both the Read and Admin roles
		// everywhere. There is no issue with a role appearing twice in this list.
		cfg.Permissions.Read = append(strings.Split(readRoles, ","), cfg.Permissions.Admin...)
	}

	// Load write IAM roles
	if writeRoles := os.Getenv("WRITE_ACCESS_ROLES"); writeRoles != "" {
		// As admins are allowed to perform any action we append their roles to the
		// Write list of ARNs to prevent having to add both the Write and Admin roles
		// everywhere. There is no issue with a role appearing twice in this list.
		cfg.Permissions.Write = append(strings.Split(writeRoles, ","), cfg.Permissions.Admin...)
	}

	return cfg
}
