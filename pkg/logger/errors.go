package logger

import "errors"

var (
	ErrorUnknownLevel = errors.New("unknown level")
	ErrorUnknownMode  = errors.New("unknown mode")
)
