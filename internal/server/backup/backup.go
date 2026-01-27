package backup

import (
	"context"
	"time"

	"github.com/mrPTqp/metrics/internal/server/service"
	"github.com/mrPTqp/metrics/internal/server/storage"
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

func (b *Backuper) Backup(ctx context.Context) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	gauges, counters := b.service.ListAllMetrics(timeoutCtx)

	err := b.storage.SaveAllMetrics(timeoutCtx, gauges, counters)
	if err != nil {
		b.logger.Errorf("Failed to save metrics: %v", err)
		return err
	}

	b.logger.Infoln("Metrics saved successfully")
	return nil
}
