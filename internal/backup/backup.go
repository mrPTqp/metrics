package backup

import (
	"github.com/mrPTqp/metrics/internal/service"
	"github.com/mrPTqp/metrics/internal/storage"
	"go.uber.org/zap"
)

type Backuper struct {
	service service.MetricsService
	storage *storage.FileStorage
	logger  *zap.SugaredLogger
}

func NewBackuper(service service.MetricsService, storage *storage.FileStorage, logger *zap.SugaredLogger) *Backuper {
	return &Backuper{
		service: service,
		storage: storage,
		logger:  logger,
	}
}

func (b *Backuper) Backup() {
	gauges, counters := b.service.ListAllMetrics()

	err := b.storage.SaveAllMetrics(gauges, counters)
	if err != nil {
		b.logger.Errorf("Failed to save metrics: %v", err)
	} else {
		b.logger.Info("Metrics saved successfully")
	}
}