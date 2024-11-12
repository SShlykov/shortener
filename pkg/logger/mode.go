package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/sshlykov/shortener/pkg/logger/handler"
)

type Mode int

const (
	ModePretty Mode = iota
	ModeJSON
)

func modeToHandler(mode Mode, opts handler.Options) (slog.Handler, error) {
	switch mode {
	case ModePretty:
		return handler.NewHandler(opts), nil
	case ModeJSON:
		return slog.NewJSONHandler(os.Stdout, &opts.SlogOpts), nil
	default:
		return nil, ErrorUnknownMode
	}
}

func ModeFromString(s string) (Mode, error) {
	switch strings.ToLower(s) {
	case "pretty":
		return ModePretty, nil
	case "json":
		return ModeJSON, nil
	default:
		return 0, ErrorUnknownMode
	}
}
