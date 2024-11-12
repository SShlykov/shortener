package app

import "errors"

var (
	ErrCantStart       = errors.New("can't start app")
	ErrTimeoutExceeded = errors.New("timeout exceeded")
)
