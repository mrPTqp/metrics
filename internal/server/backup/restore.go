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

type Restorer struct {
	service service.MetricsService
	storage *storage.FileStorage
	logger  *zap.Logger
}

func NewRestorer(service service.MetricsService, storage *storage.FileStorage, logger *zap.Logger) *Restorer {
	return &Restorer{
		service: service,
		storage: storage,
		logger:  logger,
	}
}

func (r *Restorer) Restore(ctx context.Context) error {
	log := contextkey.LoggerFromContext(ctx)
	restoreCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	gauges, counters, err := r.storage.Restore()
	if err != nil {
		var syntaxError *json.SyntaxError
		if errors.Is(err, io.EOF) {
			log.Info("Snapshot file is empty, starting fresh")
			return nil
		}
		if errors.As(err, &syntaxError) {
			log.Warn("Snapshot file is corrupted, starting fresh", zap.Error(err))
			return nil
		}
		log.Fatal("Failed to restore data", zap.Error(err))
	}

	if len(gauges) == 0 && len(counters) == 0 {
		log.Info("No metrics found in snapshot, starting fresh")
		return nil
	}

	if err := r.service.SaveAllMetrics(restoreCtx, gauges, counters); err != nil {
		log.Error("Failed to load metrics into memory", zap.Error(err))
		return err
	}

	log.Info("Restore completed",
		zap.Int("gauges", len(gauges)),
		zap.Int("counters", len(counters)))
	return nil
}
