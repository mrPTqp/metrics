package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/server/repository"
)

// Декоратор для бэкапа метрик в файл
type FileBackupService struct {
	service MetricsService
	repo    repository.MetricRepository
	logger  *zap.Logger
}

// Возвращает новый экземпляр FileBackupService
func NewFileBackupService(service MetricsService, repo repository.MetricRepository, logger *zap.Logger) *FileBackupService {
	return &FileBackupService{
		service: service,
		repo:    repo,
		logger:  logger,
	}
}

// Прокидывает запросы в базовый сервис и сохраняет gauge метрику в файл
func (s *FileBackupService) SaveGaugeMetric(ctx context.Context, mName string, mValue *float64) error {
	log := contextkey.LoggerFromContext(ctx)
	if err := s.service.SaveGaugeMetric(ctx, mName, mValue); err != nil {
		return err
	}
	if err := s.repo.SaveGauge(ctx, mName, mValue); err != nil {
		log.Warn("Failed to backup gauge to file", zap.String("name", mName), zap.Error(err))
	}
	return nil
}

// Прокидывает запросы в базовый сервис и сохраняет counterметрику в файл
func (s *FileBackupService) SaveCounterMetric(ctx context.Context, mName string, mValue *int64) error {
	log := contextkey.LoggerFromContext(ctx)
	if err := s.service.SaveCounterMetric(ctx, mName, mValue); err != nil {
		return err
	}
	if err := s.repo.SaveCounter(ctx, mName, mValue); err != nil {
		log.Warn("Failed to backup counter to file", zap.String("name", mName), zap.Error(err))
	}
	return nil
}

// Прокси
func (s *FileBackupService) GetGaugeMetric(ctx context.Context, mName string) (float64, error) {
	return s.service.GetGaugeMetric(ctx, mName)
}

// Прокси
func (s *FileBackupService) GetCounterMetric(ctx context.Context, mName string) (int64, error) {
	return s.service.GetCounterMetric(ctx, mName)
}

// Прокси
func (s *FileBackupService) ListAllMetrics(ctx context.Context) (map[string]float64, map[string]int64) {
	return s.service.ListAllMetrics(ctx)
}

// Прокидывает запросы в базовый сервис и сохраняет gauge и counter метрики в файл
func (s *FileBackupService) SaveAllMetrics(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	if err := s.service.SaveAllMetrics(ctx, gauges, counters); err != nil {
		return err
	}
	return s.repo.SaveAllMetrics(ctx, gauges, counters)
}

// Прокси
func (s *FileBackupService) Ping(ctx context.Context) bool {
	return s.repo.CheckStorageAvailability(ctx)
}
