package registry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"

	healthcntrl "github.com/sshlykov/shortener/internal/app/health"
	"github.com/sshlykov/shortener/internal/config"
	"github.com/sshlykov/shortener/pkg/logger"
	mw "github.com/sshlykov/shortener/pkg/logger/echomw"
)

func RunHealthServer(ctx context.Context, prom *prometheus.Registry, cfg config.Health,
	readinessHandler func() bool) error {

	handler := echo.New()
	handler.Use(middleware.Recover())
	// metrics middleware
	// tracer middleware

	loggermw := mw.New(*logger.FromContext(ctx))
	handler.Use(loggermw)

	healthcntrl.New(prom, readinessHandler).RegisterRoutes(handler.Group(""))
	//healthcntrl.New(prom, readinessHandler).RegisterRoutes(handler.Group("private")) + middleware.BasicAuth

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	go func() {
		logger.Info(ctx, "Health server started", logger.Any("address", server.Addr))
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error(ctx, "error", logger.Err(err))
		}
	}()
	return handleHealthClose(ctx, server, cfg.ShutdownTimeout)
}

func handleHealthClose(ctx context.Context, server *http.Server, shutdownTimeout time.Duration) error {
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// тут graceful shutdown, поэтому искусственно создаем новый контекст (линтер считает, что это ошибка.
	// тк. создается новый контекст, а не модифицируется исходный)
	//nolint:contextcheck
	return server.Shutdown(shutdownCtx)
}
