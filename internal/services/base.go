package services

import (
	"github.com/sshlykov/shortener/internal/bootstrap/metrics"
	"github.com/sshlykov/shortener/internal/config"
	"github.com/sshlykov/shortener/pkg/logger"
)

type BaseService struct {
	name    string
	cfg     config.BaseServiceConfig
	logger  logger.Logger
	metrics metrics.MetricsCollector
}

func (s *BaseService) Name() string {
	return s.name
}
