package services

import (
	"context"

	"github.com/sshlykov/shortener/internal/config"
)

type OrderService struct {
	BaseService
	processor OrderProcessor
}

func NewOrderService(cfg config.OrderServiceConfig) (*OrderService, error) {
	baseService := BaseService{
		name: "order_service",
		cfg: config.BaseServiceConfig{
			LogLevel: cfg.LogLevel,
			Enabled:  cfg.Enabled,
		},
	}
	return &OrderService{
		BaseService: baseService,
		processor:   NewOrderProcessor(),
	}, nil
}

func (s *OrderService) Start(ctx context.Context) error {
	// Инициализация и запуск сервиса
	return nil
}

func (s *OrderService) Stop(ctx context.Context) error {
	// Graceful shutdown сервиса
	return nil
}
