package backup

import (
	"context"
	"time"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/server/service"
	"github.com/mrPTqp/metrics/internal/server/storage"
	"go.uber.org/zap"
)

type Backuper struct {
	service service.MetricsService
	storage *storage.FileStorage
	logger  *zap.Logger
}

func NewBackuper(service service.MetricsService, storage *storage.FileStorage, logger *zap.Logger) *Backuper {
	return &Backuper{
		service: service,
		storage: storage,
		logger:  logger,
	}
}

func (b *Backuper) Backup(ctx context.Context) error {
	log := contextkey.LoggerFromContext(ctx)
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	gauges, counters := b.service.ListAllMetrics(timeoutCtx)

	err := b.storage.SaveAllMetrics(timeoutCtx, gauges, counters)
	if err != nil {
		log.Error("Failed to save metrics to file", zap.Error(err))
		return err
	}

	log.Info("Metrics successfully backed up")
	return nil
}
