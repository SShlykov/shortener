package service

import (
	"context"
	"time"

	"github.com/sshlykov/shortener/pkg/logger"
)

func (s *Service) SelectNow(ctx context.Context) (*time.Time, error) {
	res, err := s.repo.SelectNow(ctx)
	if err != nil {
		logger.Error(ctx, "SelectNow", logger.Err(err))

		return nil, ErrCantGetNow
	}

	timeRes, ok := res.(time.Time)
	if !ok {
		logger.Error(ctx,
			"SelectNow",
			logger.Err(ErrInvalidResultType),
			logger.Any("result", res),
		)

		return nil, ErrInvalidResultType
	}

	return &timeRes, nil
}
