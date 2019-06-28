package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	logger "github.com/lob/logger-go"
	metrics "github.com/lob/metrics-go"
	"github.com/lob/pharos/pkg/pharos-api-server/application"
	"github.com/lob/pharos/pkg/pharos-api-server/binder"
	"github.com/lob/pharos/pkg/pharos-api-server/clusters"
	"github.com/lob/pharos/pkg/pharos-api-server/health"
	"github.com/lob/pharos/pkg/pharos-api-server/recovery"
	"github.com/lob/pharos/pkg/pharos-api-server/signals"
	sentryecho "github.com/lob/sentry-echo/pkg"
)

// New returns a new HTTP server with the registered routes.
func New(app application.App) *http.Server {
	e := echo.New()
	log := logger.New()

	b := binder.New()
	e.Binder = b

	e.Use(metrics.Middleware(app.Metrics))
	e.Use(logger.Middleware())
	e.Use(recovery.Middleware())

	sentryecho.RegisterErrorHandlerWithOptions(e, sentryecho.Options{
		Reporter:                  &app.Sentry,
		EnableCustomErrorMessages: true,
	})

	health.RegisterRoutes(e)
	clusters.RegisterRoutes(e, app)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.Config.Port),
		Handler:      e,
		ReadTimeout:  65 * time.Second,
		WriteTimeout: 65 * time.Second,
	}

	graceful := signals.Setup()

	go func() {
		<-graceful
		err := srv.Shutdown(context.Background())
		if err != nil {
			log.Err(err).Error("server shutdown")
		}
	}()

	return srv
}
