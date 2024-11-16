package app

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sshlykov/shortener/internal/bootstrap/metrics"
	"github.com/sshlykov/shortener/pkg/logger"
)

func (a *App) initLogger() error {
	log, err := logger.Setup(logger.Level(a.cfg.Logger.Level), logger.Mode(a.cfg.Logger.Mode))
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	a.logger = log
	return nil
}

func (a *App) initMetrics() error {
	reg := prometheus.NewRegistry()
	a.metrics = metrics.NewMetricsCollector(reg)
	return nil
}

/*
	func (a *App) initDatabase() error {
		if !a.cfg.Database.Enabled {
			a.logger.Info("database is disabled")
			return nil
		}

		db, err := database.New(a.cfg.Database)
		if err != nil {
			return fmt.Errorf("failed to create database client: %w", err)
		}

		// Проверяем подключение
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			return fmt.Errorf("failed to ping database: %w", err)
		}

		a.db = db
		return nil
	}

	func (a *App) initCache() error {
		if !a.cfg.Cache.Enabled {
			a.logger.Info("cache is disabled")
			return nil
		}

		cache, err := cache.New(a.cfg.Cache)
		if err != nil {
			return fmt.Errorf("failed to create cache client: %w", err)
		}

		// Проверяем подключение
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := cache.Ping(ctx); err != nil {
			return fmt.Errorf("failed to ping cache: %w", err)
		}

		a.cache = cache
		return nil
	}

	func (a *App) initEventBus() error {
		if !a.cfg.EventBus.Enabled {
			a.logger.Info("event bus is disabled")
			return nil
		}

		eventBus, err := eventbus.New(a.cfg.EventBus)
		if err != nil {
			return fmt.Errorf("failed to create event bus client: %w", err)
		}

		a.eventBus = eventBus
		return nil
	}

	func (a *App) initExternalClients() error {
		// Инициализируем платежный шлюз
		if a.cfg.Payment.Enabled {
			paymentClient, err := payment.New(a.cfg.Payment)
			if err != nil {
				return fmt.Errorf("failed to create payment client: %w", err)
			}
			a.paymentGateway = paymentClient
		}

		// Инициализируем email провайдер
		if a.cfg.Email.Enabled {
			emailClient, err := email.New(a.cfg.Email)
			if err != nil {
				return fmt.Errorf("failed to create email client: %w", err)
			}
			a.emailProvider = emailClient
		}

		return nil
	}
*/
func (a *App) initHealthCheck() error {
	health, err := NewHealthCheck(a.cfg.Health, a.metrics)
	if err != nil {
		return fmt.Errorf("failed to create health check: %w", err)
	}

	// Добавляем проверки для компонентов
	/*if a.db != nil {
		health.AddChecker(NewDatabaseChecker(a.db))
	}
	if a.cache != nil {
		health.AddChecker(NewCacheChecker(a.cache))
	}
	if a.eventBus != nil {
		health.AddChecker(NewEventBusChecker(a.eventBus))
	}
	if a.paymentGateway != nil {
		health.AddChecker(NewPaymentChecker(a.paymentGateway))
	}*/

	a.healthCheck = health
	return nil
}

func (a *App) initServices() error {
	// Собираем все зависимости для сервисов
	deps := ServiceDependencies{
		Logger:  a.logger,
		Metrics: a.metrics,
		/*Database: a.db,
		Cache:    a.cache,
		EventBus: a.eventBus,
		ExternalClients: ExternalClients{
			Payment: a.paymentGateway,
			Email:   a.emailProvider,
		},*/
	}

	// Создаем сервисы
	svcs, err := BuildServices(a.cfg.Services, deps)
	if err != nil {
		return fmt.Errorf("failed to build services: %w", err)
	}

	a.services = svcs
	return nil
}
