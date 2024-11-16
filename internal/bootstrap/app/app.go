package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/sshlykov/shortener/internal/config"
	"github.com/sshlykov/shortener/pkg/logger"
)

type App struct {
	cfg         config.AppConfig
	logger      Logger
	metrics     MetricsCollector
	healthCheck HealthChecker
	services    []Service
	//db          *database.Client
	//cache       *cache.Client
	//eventBus    *eventbus.Client

	// Внешние клиенты
	//paymentGateway *payment.Client
	//emailProvider  *email.Client
}

func New(cfg config.AppConfig) (*App, error) {
	app := &App{
		cfg: cfg,
	}

	initSteps := []struct {
		name string
		fn   func() error
	}{
		{"logger", app.initLogger},
		{"metrics", app.initMetrics},
		{"health check", app.initHealthCheck},
		{"services", app.initServices},
		//{"database", app.initDatabase},
		//{"cache", app.initCache},
		//{"event bus", app.initEventBus},
		//{"external clients", app.initExternalClients},
	}

	for _, step := range initSteps {
		if err := step.fn(); err != nil {
			return nil, fmt.Errorf("failed to initialize %s: %w", step.name, err)
		}
	}

	return app, nil
}

func (a *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errChan := make(chan error, len(a.services))

	for _, svc := range a.services {
		go func(s Service) {
			logger.Info(ctx, "starting service", "name", s.Name())
			if err := s.Start(ctx); err != nil {
				errChan <- fmt.Errorf("service %s error: %w", s.Name(), err)
			}
		}(svc)
	}

	go func() {
		err := a.healthCheck.Start(ctx)
		if err != nil {
			errChan <- fmt.Errorf("health check error: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		return a.shutdown(context.Background())
	case err := <-errChan:
		return fmt.Errorf("service error: %w", err)
	}
}

func (a *App) stopServices(ctx context.Context) error {
	var errs []error

	for _, svc := range a.services {
		if err := svc.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop %s service: %w", svc.Name(), err))
		}
	}

	return errors.Join(errs...)
}

func (a *App) shutdown(ctx context.Context) error {
	var errs []error

	if err := a.stopServices(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to stop services: %w", err))
	}

	if a.healthCheck != nil {
		if err := a.healthCheck.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop health check: %w", err))
		}
	}

	// Закрываем соединения с внешними сервисами
	/*
		if a.paymentGateway != nil {
			if err := a.paymentGateway.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close payment gateway: %w", err))
			}
		}

		if a.emailProvider != nil {
			if err := a.emailProvider.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close email provider: %w", err))
			}
		}

		// Закрываем внутренние компоненты
		if a.eventBus != nil {
			if err := a.eventBus.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close event bus: %w", err))
			}
		}

		if a.cache != nil {
			if err := a.cache.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close cache: %w", err))
			}
		}

		if a.db != nil {
			if err := a.db.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close database: %w", err))
			}
		}

	*/

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
