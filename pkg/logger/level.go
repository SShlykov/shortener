package logger

import (
	"log/slog"
)

type Level int

const (
	LevelDebug Level = 4*iota - 4
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) Level() slog.Level {
	return slog.Level(l)
}

func LevelFromString(s string) (Level, error) {
	switch s {
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	default:
		return 0, ErrorUnknownLevel
	}
}
