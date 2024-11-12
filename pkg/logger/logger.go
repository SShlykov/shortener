package logger

import (
	"context"
	"log/slog"

	"github.com/sshlykov/shortener/pkg/logger/handler"
)

type Logger struct {
	logger *slog.Logger
}

// Setup создает новый логгер с заданными параметрами
//
// level - нижний уровень логирования (debug, info, warn, error);
// mode  - режим логирования (pretty);
func Setup(level Level, mode Mode) (*Logger, error) {
	opts := handler.Options{
		SlogOpts: slog.HandlerOptions{
			Level: level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				switch a.Key {
				case slog.LevelKey:
					a.Key = "level"
				case slog.MessageKey:
					a.Key = "message"
				case slog.TimeKey:
					a.Key = "@timestamp"
					return a
				default:
				}
				return a
			},
		},
	}

	h, err := modeToHandler(mode, opts)
	if err != nil {
		return nil, err
	}
	lggr := slog.New(h)

	return &Logger{logger: lggr}, nil
}

func (l *Logger) Extract() *slog.Logger {
	return l.logger
}

func (l *Logger) With(attrs ...any) {
	l.logger = l.logger.With(attrs...)
}

func (l *Logger) WithGroup(name string) {
	l.logger = l.logger.WithGroup(name)
}

func (l *Logger) LogAttrs(ctx context.Context, level Level, msg string, attrs ...slog.Attr) {
	l.logger.LogAttrs(ctx, level.Level(), msg, attrs...)
}

// Методы логгера сделаны стаким образом, чтобы можно было использовать в качестве логгера для krakend и при этом
// не терять старые логгеры (структура любого лога всегда будет содержать строку и набор интерфейсов)
// К сожалению в krakend нет возможности использовать slog.Logger напрямую

func (l *Logger) Warn(attrs ...interface{}) {
	l.logger.Warn(attrs[0].(string), attrs[1:]...)
}

func (l *Logger) Warning(attrs ...interface{}) {
	l.logger.Warn(attrs[0].(string), attrs[1:]...)
}

func (l *Logger) Info(attrs ...interface{}) {
	l.logger.Info(attrs[0].(string), attrs[1:]...)
}

func (l *Logger) Debug(attrs ...interface{}) {
	l.logger.Debug(attrs[0].(string), attrs[1:]...)
}

func (l *Logger) Error(attrs ...interface{}) {
	l.logger.Error(attrs[0].(string), attrs[1:]...)
}

func (l *Logger) Critical(attrs ...interface{}) {
	l.logger.Error(attrs[0].(string), attrs[1:]...)
}

func (l *Logger) Fatal(attrs ...interface{}) {
	l.logger.Error(attrs[0].(string), attrs[1:]...)
}
