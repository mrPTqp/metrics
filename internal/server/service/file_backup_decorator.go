package service

import (
	"context"
	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/server/repository"
)

type FileBackupService struct {
	service MetricsService
	repo    repository.MetricRepository
	logger  *zap.Logger
}

func NewFileBackupService(service MetricsService, repo repository.MetricRepository, logger *zap.Logger) *FileBackupService {
	return &FileBackupService{
		service: service,
		repo:    repo,
		logger:  logger,
	}
}

func (d *FileBackupService) SaveGaugeMetric(ctx context.Context, mName string, mValue *float64) error {
	log := contextkey.LoggerFromContext(ctx)
	if err := d.service.SaveGaugeMetric(ctx, mName, mValue); err != nil {
		return err
	}
	if err := d.repo.SaveGauge(ctx, mName, mValue); err != nil {
		log.Warn("Failed to backup gauge to file", zap.String("name", mName), zap.Error(err))
	}
	return nil
}

func (d *FileBackupService) SaveCounterMetric(ctx context.Context, mName string, mValue *int64) error {
	log := contextkey.LoggerFromContext(ctx)
	if err := d.service.SaveCounterMetric(ctx, mName, mValue); err != nil {
		return err
	}
	if err := d.repo.SaveCounter(ctx, mName, mValue); err != nil {
		log.Warn("Failed to backup counter to file", zap.String("name", mName), zap.Error(err))
	}
	return nil
}

func (d *FileBackupService) GetGaugeMetric(ctx context.Context, mName string) (float64, error) {
	return d.service.GetGaugeMetric(ctx, mName)
}

func (d *FileBackupService) GetCounterMetric(ctx context.Context, mName string) (int64, error) {
	return d.service.GetCounterMetric(ctx, mName)
}

func (d *FileBackupService) ListAllMetrics(ctx context.Context) (map[string]float64, map[string]int64) {
	return d.service.ListAllMetrics(ctx)
}

func (d *FileBackupService) SaveAllMetrics(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	if err := d.service.SaveAllMetrics(ctx, gauges, counters); err != nil {
		return err
	}
	return d.repo.SaveAllMetrics(ctx, gauges, counters)
}

func (d *FileBackupService) Ping(ctx context.Context) bool {
	return d.repo.CheckStorageAvailability(ctx)
}
