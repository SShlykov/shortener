package registry

import (
	"context"

	"github.com/sshlykov/shortener/internal/config"
)

type Services struct {
	HealthService
	// поворот картинки
	// масштабирование
	// обрезка
	// сервисный слой приложения
	// user service
	// role service
}

type HealthService interface {
	Test(context.Context) string
}

func NewServices(_ *config.Config) *Services {
	var health HealthService

	return &Services{
		HealthService: health,
	}
}
