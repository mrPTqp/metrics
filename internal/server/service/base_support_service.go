package service

import (
	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/server/repository"
)

type BaseSupportService struct {
	repo   repository.MetricRepository
	logger *zap.SugaredLogger
}

func NewBaseSupportService(repo repository.MetricRepository, logger *zap.SugaredLogger) *BaseSupportService {
	return &BaseSupportService{
		repo:   repo,
		logger: logger,
	}
}


