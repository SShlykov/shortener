package registry

import (
	"context"
	"net/http"
	"time"
)

func handleHTTPClose(ctx context.Context, server *http.Server, shutdownTimeout time.Duration) error {
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// тут graceful shutdown, поэтому искусственно создаем новый контекст (линтер считает, что это ошибка.
	// тк. создается новый контекст, а не модифицируется исходный)
	//nolint:contextcheck
	return server.Shutdown(shutdownCtx)
}
