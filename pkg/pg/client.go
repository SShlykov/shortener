package pg

import (
	"time"
)

const (
	_defaultMaxPoolSize  = 10
	_defaultConnAttempts = 1
	_defaultConnTimeout  = time.Second
)

type pgClient struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	//	db - should be your db client interface
}

func NewClient() (db any, err error) {

	panic("implement me")
}
