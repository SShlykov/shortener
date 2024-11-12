package app

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sshlykov/shortener/internal/bootstrap/registry"
	"github.com/sshlykov/shortener/internal/config"
	"github.com/sshlykov/shortener/pkg/logger"
	"github.com/sshlykov/shortener/pkg/postgres"
	"go.opentelemetry.io/otel/sdk/trace"
)

type App struct {
	//nolint:containedctx
	ctx    context.Context
	cancel context.CancelFunc

	db postgres.Client

	cfg  *config.Config
	prom *prometheus.Registry

	traceProvider *trace.TracerProvider

	ready    int32
	checkers []DependencyChecker

	services *registry.Services
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	app := &App{}
	app.ctx, app.cancel = context.WithCancel(ctx)
	app.cfg = cfg

	for _, init := range app.appInitials() {
		if err := init(); err != nil {
			return nil, err
		}
	}

	return app, nil
}

func (app *App) Run() (err error) {
	ctx, stop := signal.NotifyContext(app.ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	logger.Info(ctx, "starting app")
	logger.Debug(ctx, "debug messages started")

	app.services = registry.NewServices(app.cfg)

	for _, checker := range app.appCheckers() {
		app.RegisterChecker(checker)
	}

	for _, service := range app.appServices() {
		wg.Add(1)
		go service(ctx, app.cancel, &wg)
	}

	stoppedChan := make(chan struct{})
	go func() {
		wg.Wait()
		stoppedChan <- struct{}{}
	}()

	return app.closer(ctx, stoppedChan)
}
