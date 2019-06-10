package application

import (
	"github.com/go-pg/pg"
	"github.com/lob/metrics-go"
	"github.com/lob/pharos/pkg/pharos-api-server/config"
	"github.com/lob/pharos/pkg/pharos-api-server/database"
	"github.com/lob/sentry-echo/pkg/sentry"
	"github.com/pkg/errors"
)

// App contains necessary references that will be persisted throughout the
// application's lifecycle.
type App struct {
	Config  config.Config
	DB      *pg.DB
	Metrics metrics.Metrics
	Sentry  sentry.Sentry
}

// New creates a new instance of App with Config, DB connection, Metrics and Sentry.
func New() (App, error) {
	cfg := config.New()

	db, err := database.New(cfg)
	if err != nil {
		return App{}, err
	}

	m, err := metrics.New(metrics.Config{
		Environment: cfg.Environment,
		Hostname:    cfg.Hostname,
		Namespace:   "pharos-api-server",
		StatsdHost:  cfg.StatsdHost,
		StatsdPort:  cfg.StatsdPort,
	})
	if err != nil {
		return App{}, errors.Wrap(err, "application")
	}

	s, err := sentry.New(cfg.SentryDSN)
	if err != nil {
		return App{}, errors.Wrap(err, "application")
	}

	return App{cfg, db, m, s}, nil
}
