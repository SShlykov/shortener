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
		app.runWebApp,
		app.runHealthApp,
		app.runReadinessChecker,
	}
}

func (app *App) runWebApp(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	defer stop()
	defer logger.Info(ctx, "web app stopped")

	if err := registry.RunWebServer(app.ctx, app.prom, app.cfg.Web, app.services); err != nil {
		logger.Error(ctx, "web app error", err)
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
