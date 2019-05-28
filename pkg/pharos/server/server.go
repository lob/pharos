package server

// New returns a new HTTP server with the registered routes.
import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	logger "github.com/lob/logger-go"
	metrics "github.com/lob/metrics-go"
	"github.com/lob/pharos/pkg/pharos/application"
	"github.com/lob/pharos/pkg/pharos/health"
	"github.com/lob/pharos/pkg/pharos/recovery"
	"github.com/lob/pharos/pkg/pharos/signals"
	sentryecho "github.com/lob/sentry-echo/pkg"
)

// New returns a new HTTP server with the registered routes.
func New(app application.App) *http.Server {
	e := echo.New()
	log := logger.New()

	e.Use(metrics.Middleware(app.Metrics))
	e.Use(logger.Middleware())
	e.Use(recovery.Middleware())

	sentryecho.RegisterErrorHandler(e, &app.Sentry)

	health.RegisterRoutes(e)

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
