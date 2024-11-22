package service

import (
	"context"

	"github.com/sshlykov/shortener/pkg/postgres"

	repository "github.com/sshlykov/shortener/internal/pkg/test_feat/repo"
)

type Service struct {
	repo Repository
}

type Repository interface {
	SelectNow(ctx context.Context) (interface{}, error)
}

func New(db postgres.Client) *Service {
	repo := repository.New(db)

	return &Service{
		repo: repo,
	}
}
