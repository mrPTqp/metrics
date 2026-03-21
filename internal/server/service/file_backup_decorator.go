package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/server/repository"
)

// Декоратор для бэкапа метрик в файл
type FileBackupService struct {
	writer MetricWriter
	repo   repository.MetricRepository
	logger *zap.Logger
}

// Возвращает новый экземпляр FileBackupService
func NewFileBackupService(writer MetricWriter, repo repository.MetricRepository, logger *zap.Logger) *FileBackupService {
	return &FileBackupService{
		writer: writer,
		repo:   repo,
		logger: logger,
	}
}

// Прокидывает запросы в базовый сервис и сохраняет gauge метрику в файл
func (s *FileBackupService) SaveGaugeMetric(ctx context.Context, mName string, mValue *float64) error {
	log := contextkey.LoggerFromContext(ctx)
	if err := s.writer.SaveGaugeMetric(ctx, mName, mValue); err != nil {
		return err
	}
	if err := s.repo.SaveGauge(ctx, mName, mValue); err != nil {
		log.Warn("failed to backup gauge to file", zap.String("name", mName), zap.Error(err))
	}
	return nil
}

// Прокидывает запросы в базовый сервис и сохраняет counterметрику в файл
func (s *FileBackupService) SaveCounterMetric(ctx context.Context, mName string, mValue *int64) error {
	log := contextkey.LoggerFromContext(ctx)
	if err := s.writer.SaveCounterMetric(ctx, mName, mValue); err != nil {
		return err
	}
	if err := s.repo.SaveCounter(ctx, mName, mValue); err != nil {
		log.Warn("failed to backup counter to file", zap.String("name", mName), zap.Error(err))
	}
	return nil
}

// Прокидывает запросы в базовый сервис и сохраняет gauge и counter метрики в файл
func (s *FileBackupService) SaveAllMetrics(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	if err := s.writer.SaveAllMetrics(ctx, gauges, counters); err != nil {
		return err
	}
	return s.repo.SaveAllMetrics(ctx, gauges, counters)
}
