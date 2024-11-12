package app

import (
	"context"
	"sync/atomic"

	"github.com/sshlykov/shortener/pkg/logger"
)

func (app *App) appCheckers() []DependencyChecker {
	return []DependencyChecker{
		//	checkers.NewDBChecker(app.db),
		//	checkers.NewAPIChecker(fmt.Sprintf("http://localhost:%d/health", app.cfg.Health.Port)),
	}
}

func (app *App) RegisterChecker(checker DependencyChecker) {
	app.checkers = append(app.checkers, checker)
}

func (app *App) CheckReadiness(ctx context.Context) bool {
	for _, checker := range app.checkers {
		if err := checker.Check(ctx); err != nil {
			logger.Error(ctx, "Readiness check failed", logger.Err(err))
			atomic.StoreInt32(&app.ready, 0)
			return false
		}
	}

	atomic.StoreInt32(&app.ready, 1)
	return true
}

func (app *App) IsReady() bool {
	return atomic.LoadInt32(&app.ready) == 1
}
