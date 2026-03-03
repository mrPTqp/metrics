package backup

import (
	"context"
	"time"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/server/service"
	"github.com/mrPTqp/metrics/internal/server/storage"
	"go.uber.org/zap"
)

// Создатель резервной копии метрик
type Backuper struct {
	lister  service.MetricsLister
	storage *storage.FileStorage
	logger  *zap.Logger
}

// Возвращает новый экземпляр Backuper
func NewBackuper(writer service.MetricWriter, lister service.MetricsLister, storage *storage.FileStorage, logger *zap.Logger) *Backuper {
	return &Backuper{
		lister: lister,
		storage: storage,
		logger:  logger,
	}
}

// Выполняет резервное копирование метрик в файл
func (b *Backuper) Backup(ctx context.Context) error {
	log := contextkey.LoggerFromContext(ctx)
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	gauges, counters := b.lister.ListAllMetrics(timeoutCtx)

	err := b.storage.SaveAllMetrics(timeoutCtx, gauges, counters)
	if err != nil {
		log.Error("failed to save metrics to file", zap.Error(err))
		return err
	}

	log.Info("metrics successfully backed up")
	return nil
}
