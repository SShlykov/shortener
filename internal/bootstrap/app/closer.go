package app

import (
	"context"
	"time"
)

func (app *App) closer(ctx context.Context, stoppedChan <-chan struct{}) error {
	<-ctx.Done()

	// взять таймаут из конфига
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case <-stoppedChan:
		return nil
	case <-timeoutCtx.Done():
		//logger.Error()

		return nil
	}
}
