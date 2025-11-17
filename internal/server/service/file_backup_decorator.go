package service

import (
	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/server/repository"
)

type FileBackupService struct {
	service MetricsService
	repo    repository.MetricRepository
	logger  *zap.SugaredLogger
}

func NewFileBackupService(service MetricsService, repo repository.MetricRepository, logger *zap.SugaredLogger) *FileBackupService {
	return &FileBackupService{
		service: service,
		repo:    repo,
		logger:  logger,
	}
}

func (d *FileBackupService) SaveGaugeMetric(mName string, mValue *float64) error {
	if err := d.service.SaveGaugeMetric(mName, mValue); err != nil {
		return err
	}
	if err := d.repo.SaveGauge(mName, mValue); err != nil {
		d.logger.Warnw("Failed to backup gauge to file", "name", mName, "error", err)
	}
	return nil
}

func (d *FileBackupService) SaveCounterMetric(mName string, mValue *int64) error {
	if err := d.service.SaveCounterMetric(mName, mValue); err != nil {
		return err
	}
	if err := d.repo.SaveCounter(mName, mValue); err != nil {
		d.logger.Warnw("Failed to backup counter to file", "name", mName, "error", err)
	}
	return nil
}

func (d *FileBackupService) GetGaugeMetric(mName string) (float64, error) {
	return d.service.GetGaugeMetric(mName)
}

func (d *FileBackupService) GetCounterMetric(mName string) (int64, error) {
	return d.service.GetCounterMetric(mName)
}

func (d *FileBackupService) ListAllMetrics() (map[string]float64, map[string]int64) {
	return d.service.ListAllMetrics()
}

func (d *FileBackupService) SaveAllMetrics(gauges map[string]float64, counters map[string]int64) error {
	if err := d.service.SaveAllMetrics(gauges, counters); err != nil {
		return err
	}
	return d.repo.SaveAllMetrics(gauges, counters)
}

func (d *FileBackupService) Ping() bool {
	return d.repo.CheckStorageAvailability()
}
