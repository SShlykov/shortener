package services

import (
	"context"
	"fmt"

	"github.com/sshlykov/shortener/internal/config"
)

type UserService struct {
	BaseService
	db Database
	//cache  Cache
	//events EventPublisher
}

func NewUserService(cfg config.UserServiceConfig) (*UserService, error) {
	db, err := NewDatabase(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	/*
		cache, err := NewCache(cfg.Cache)
		if err != nil {
			return nil, fmt.Errorf("failed to create cache: %w", err)
		}

		events, err := NewEventPublisher(cfg.Events)
		if err != nil {
			return nil, fmt.Errorf("failed to create event publisher: %w", err)
		}
	*/
	return &UserService{
		BaseService: BaseService{
			name: "user_service",
			cfg: config.BaseServiceConfig{
				LogLevel: cfg.LogLevel,
				Enabled:  cfg.Enabled,
			},
		},
		db: db,
		//cache:  cache,
		//events: events,
	}, nil
}

func (s *UserService) Start(ctx context.Context) error {
	// Инициализация и запуск сервиса
	return nil
}

func (s *UserService) Stop(ctx context.Context) error {
	// Graceful shutdown сервиса
	return nil
}
