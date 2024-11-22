package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/sshlykov/shortener/internal/config"
	"github.com/sshlykov/shortener/pkg/backoff"
	"github.com/sshlykov/shortener/pkg/logger"
	"github.com/sshlykov/shortener/pkg/postgres"
)

func (app *App) appInitials() []func() error {
	return []func() error{
		app.initLogger,
		app.initPrometheus,
		app.initOtel,
		app.initDB,
	}
}

func (app *App) initLogger() error {
	level, err := logger.LevelFromString(app.cfg.Logger.Level)
	if err != nil {
		return err
	}
	mode, err := logger.ModeFromString(app.cfg.Logger.Mode)
	if err != nil {
		return err
	}
	l, err := logger.Setup(level, mode)
	if err != nil {
		return err
	}
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	appCfg := app.cfg.App
	l.With(slog.String("inst", hostname),
		slog.String("system_version", appCfg.Version),
		slog.String("system", appCfg.Name),
		slog.String("env", appCfg.Env),
	)
	app.ctx = l.Inject(app.ctx)
	return nil
}

func (app *App) initPrometheus() error {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	app.prom = reg
	return nil
}

func (app *App) initOtel() error {
	res, err := resource.Merge(
		resource.Default(), resource.NewWithAttributes(
			semconv.SchemaURL, semconv.ServiceName(app.cfg.App.Name),
			semconv.ServiceVersion(app.cfg.App.Version), semconv.DeploymentEnvironment(app.cfg.App.Env),
		))
	if err != nil {
		return err
	}

	traceProvider := trace.NewTracerProvider(trace.WithResource(res))
	if app.cfg.App.OtelAgent != "" {
		conn, err := grpc.NewClient(app.cfg.App.OtelAgent, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		exporter, err := otlptracegrpc.New(context.Background(), otlptracegrpc.WithGRPCConn(conn))
		if err != nil {
			return err
		}
		traceProvider = trace.NewTracerProvider(
			trace.WithBatcher(exporter), trace.WithResource(res),
		)
	}
	otel.SetTracerProvider(traceProvider)
	app.traceProvider = traceProvider

	return nil
}

func (app *App) initDB() error {
	// Сделано так, чтобы дать запуститься всему остальному, если база не стартанула, то приложение завершится
	dsn, err := config.GetDSN()
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	db, err := postgres.NewClient(app.ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to init pg client: %w", err)
	}
	app.db = db

	checkupdb := func() {
		startTime := time.Now()
		h := func() error {
			if time.Since(startTime) > app.cfg.DB.RefreshTimeout {
				return backoff.Permanent(ErrTimeoutExceeded)
			}
			err = db.DB().Ping(app.ctx)
			if err == nil {
				return backoff.Permanent(nil)
			}
			return ErrCantStart
		}
		if err = backoff.Retry(h, backoff.NewExponentialBackOff()); err != nil {
			logger.Error(app.ctx, "error during db creation", logger.Err(err))
			app.cancel()
		}
	}
	go checkupdb()

	return nil
}
