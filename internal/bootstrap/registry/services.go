package registry

import (
	"context"
	"time"

	"github.com/sshlykov/shortener/internal/config"
	testsrvpkg "github.com/sshlykov/shortener/internal/pkg/test_feat/service"
	"github.com/sshlykov/shortener/pkg/postgres"
)

type Services struct {
	TestService
}

type TestService interface {
	SelectNow(ctx context.Context) (*time.Time, error)
}

func NewServices(db postgres.Client, _ *config.Config) *Services {
	testsrv := testsrvpkg.New(db)

	return &Services{
		TestService: testsrv,
	}
}
