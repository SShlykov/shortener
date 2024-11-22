package registry

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"

	webcntrl "github.com/sshlykov/shortener/internal/app/web"
	"github.com/sshlykov/shortener/internal/config"
	"github.com/sshlykov/shortener/pkg/logger"
	mw "github.com/sshlykov/shortener/pkg/logger/echomw"
)

func RunWebServer(ctx context.Context, prom *prometheus.Registry, cfg config.Web, service *Services) error {
	handler := echo.New()
	handler.Use(middleware.Recover())

	loggermw := mw.New(*logger.FromContext(ctx))
	handler.Use(loggermw)

	handler.Use(NewPrometheusMiddleware(prom).Middleware())

	webcntrl.New(service).RegisterRoutes(handler.Group(""))

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	go func() {
		logger.Info(ctx, "Web server started", logger.Any("address", server.Addr))
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error(ctx, "error", logger.Err(err))
		}
	}()
	return handleHTTPClose(ctx, server, cfg.ShutdownTimeout)
}
