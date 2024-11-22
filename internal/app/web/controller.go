package health

import (
	"context"
	"time"
)

type Service interface {
	SelectNow(ctx context.Context) (*time.Time, error)
}

type Controller struct {
	svc Service
}

func New(svc Service) *Controller {
	return &Controller{
		svc: svc,
	}
}
