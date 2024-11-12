package logger

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

const (
	loggerKey = "logger"
)

func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(loggerKey).(*Logger); ok {
		return l
	}

	return &Logger{logger: slog.Default()}
}

func (l *Logger) Inject(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func withTraces(ctx context.Context, list []any) []any {
	if spCtx := trace.SpanFromContext(ctx).SpanContext(); spCtx.IsValid() {
		list = append(list, Any("trace_id", spCtx.TraceID().String()))
		list = append(list, Any("span_id", spCtx.SpanID().String()))
	}

	return list
}

func Debug(ctx context.Context, msg string, attrs ...any) {
	if l, ok := ctx.Value(loggerKey).(*Logger); ok {
		l.logger.Debug(msg, withTraces(ctx, attrs)...)
	}
}

func Info(ctx context.Context, msg string, attrs ...any) {
	if l, ok := ctx.Value(loggerKey).(*Logger); ok {
		l.logger.Info(msg, withTraces(ctx, attrs)...)
	}
}

func Warn(ctx context.Context, msg string, attrs ...any) {
	if l, ok := ctx.Value(loggerKey).(*Logger); ok {
		l.logger.Warn(msg, withTraces(ctx, attrs)...)
	}
}

func Error(ctx context.Context, msg string, attrs ...any) {
	if l, ok := ctx.Value(loggerKey).(*Logger); ok {
		l.logger.Error(msg, withTraces(ctx, attrs)...)
	}
}
