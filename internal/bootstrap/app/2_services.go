package app

import (
	"context"
	"sync"
	"time"

	"github.com/sshlykov/shortener/internal/bootstrap/registry"
	"github.com/sshlykov/shortener/pkg/logger"
)

func (app *App) appServices() []func(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup) {
	return []func(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup){
		app.runHealthApp,
		app.runReadinessChecker,
		// app.runMetricsServer,
		// app.runTracingServer,
		// app.runAPIServer,
	}
}

func (app *App) runHealthApp(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	defer stop()
	defer logger.Info(ctx, "health app stopped")

	if err := registry.RunHealthServer(app.ctx, app.prom, app.cfg.Health, app.IsReady); err != nil {
		logger.Error(ctx, "health app error", err)
	}
}

// проверяет готовность всего приложения вместе
func (app *App) runReadinessChecker(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	defer stop()
	defer logger.Info(ctx, "Readiness checker stopped")

	ticker := time.NewTicker(app.cfg.App.ReadinessCheckPeriod)
	defer ticker.Stop()
	app.CheckReadiness(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			app.CheckReadiness(ctx)
		}
	}
}