package service

import (
	"errors"
)

var (
	ErrCantGetNow        = errors.New("can't get now")
	ErrInvalidResultType = errors.New("invalid result type")
)
