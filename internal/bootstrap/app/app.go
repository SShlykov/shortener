package app

import (
	"context"
	"os/signal"
	"runtime/metrics"
	"sync"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/sshlykov/shortener/internal/bootstrap/registry"
	"github.com/sshlykov/shortener/internal/config"
	"github.com/sshlykov/shortener/pkg/logger"
	"github.com/sshlykov/shortener/pkg/postgres"
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

// --------

func New(cfg *config.Config) *App {
	// build infrastructure
	1 / metrics
	2 / tracker
	3 / logger
	4 / db
	5 / broker - producer / consumer

	6 / readiness
	7 / health
	8 / service locator

	// build business services
	[]service
	1/ croper
	2/ zoomer
	3/ rotator
	4/ user


	return &App{
		cfg: cfg,
		registry: []goroutines / daemons,
		// others
	}
}

func Run(app *App) error {
	// goroutines runners
	1 / metrics server
	5 / broker consumer
	6 / readiness server
	7 / health server
	8 / service locator

	+ gracefull shutdown

	// business
	rest api server
	grps server
	scheduler
	parser

	// checkers
	1/ metrics health checker
	2/ db health checker
	3/ api health checker
}

// компоненты должны поддерживать интерфейсы

type Checker interface {
	Check(ctx context.Context) error
}

type Runner interface {
	Run(ctx context.Context) error
	Stop()
}

type Builder interface {
	Build(cfg *config.Config) Runner
}
