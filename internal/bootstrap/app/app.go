package app

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sshlykov/shortener/internal/config"
)

type App struct {
	ctx    context.Context
	cancel context.CancelFunc

	// db
	// traceprovider
	// prometheus
	// etc

	cfg      *config.Config
	services *registry.Services

	ready    int32
	checkers []DependencyChecker
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

	// logger.Info(ctx, "starting app")
	// logger.Debug(ctx, "starting app")

	app.services, err = registry.NewServices(app.cfg)
	if err != nil {
		return err
	}

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

		// stop everything you need
		stoppedChan <- struct{}{}
	}

	return nil
}
