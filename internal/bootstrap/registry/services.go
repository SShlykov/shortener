package health

import "context"

type Services struct {
	HealthService
}

type HealthService interface {
	Test(context.Context) string
}

func NewServices() *Services {
	var health HealthService

	return &Services{
		HealthService: health,
	}
}
