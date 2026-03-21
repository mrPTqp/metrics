package backup

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/server/service"
	"github.com/mrPTqp/metrics/internal/server/storage"
	"go.uber.org/zap"
)

// Загрузчик резервной копии метрик из файла
type Restorer struct {
	writer  service.MetricWriter
	storage *storage.FileStorage
	logger  *zap.Logger
}

// Возвращает новый экземпляр Restorer
func NewRestorer(writer service.MetricWriter, storage *storage.FileStorage, logger *zap.Logger) *Restorer {
	return &Restorer{
		writer:  writer,
		storage: storage,
		logger:  logger,
	}
}

// Загружает метрики из файловой резервной копии
func (r *Restorer) Restore(ctx context.Context) error {
	log := contextkey.LoggerFromContext(ctx)
	restoreCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	gauges, counters, err := r.storage.Restore()
	if err != nil {
		var syntaxError *json.SyntaxError
		if errors.Is(err, io.EOF) {
			log.Info("snapshot file is empty, starting fresh")
			return nil
		}
		if errors.As(err, &syntaxError) {
			log.Warn("snapshot file is corrupted, starting fresh", zap.Error(err))
			return nil
		}
		log.Fatal("failed to restore data", zap.Error(err))
	}

	if len(gauges) == 0 && len(counters) == 0 {
		log.Info("no metrics found in snapshot, starting fresh")
		return nil
	}

	if err := r.writer.SaveAllMetrics(restoreCtx, gauges, counters); err != nil {
		log.Error("failed to load metrics into memory", zap.Error(err))
		return err
	}

	log.Info("restore completed",
		zap.Int("gauges", len(gauges)),
		zap.Int("counters", len(counters)))
	return nil
}
