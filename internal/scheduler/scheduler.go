package scheduler

import (
	"time"

	"github.com/mrPTqp/metrics/internal/service"

	"go.uber.org/zap"
)

type FileBackupScheduler struct {
	service *service.BaseMetricService
	logger  *zap.SugaredLogger
}

func NewScheduler(service *service.BaseMetricService, logger *zap.SugaredLogger) *FileBackupScheduler {
	return &FileBackupScheduler{
		service: service,
		logger:  logger,
	}
}

func (fbs *FileBackupScheduler) Start(storeInterval int, fileStoragePath string) {
	ticker := time.NewTicker(time.Duration(storeInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fbs.Backup()
	}
}

func (fbs *FileBackupScheduler) Backup() {
	gauges, counters := fbs.service.ListAllMetrics()

	err := fbs.service.SaveAllMetrics(gauges, counters)
	if err != nil {
		fbs.logger.Errorf("Failed to save metrics: %v", err)
	} else {
		fbs.logger.Info("Metrics saved successfully")
	}
}
