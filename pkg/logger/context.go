package logger

import "context"

const LoggerContextKey = "logger"

func FromContext(ctx context.Context) *Logger {
	// Logger or new Logger

	return nil
}

func (l *Logger) Inject(ctx context.Context) context.Context {
	// Inject logger into context

	return ctx
}

func withTraces(ctx context.Context, list []any) []any {
	// Inject traces into context

	return nil
}

// functions Debug(ctx, msg, attrs), Info, Warn etc.
