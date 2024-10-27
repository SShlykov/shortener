package app

import (
	"context"
	"sync"
)

func (app *App) appServices() []func(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup) {
	return []func(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup){
		app.runHealthApp,
	}
}

func (app *App) runHealthApp(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	defer stop()
	// defer logger.Info(ctx, "health app stopped")

	// if err := registry.NewHealthApp(app.cfg, ...).Run(ctx); err != nil {
	// logger.Error(ctx, "health app error", err)
	// }
}
