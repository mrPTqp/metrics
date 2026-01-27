package backup

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/mrPTqp/metrics/internal/server/service"
	"github.com/mrPTqp/metrics/internal/server/storage"
	"go.uber.org/zap"
)

type Restorer struct {
	service service.MetricsService
	storage *storage.FileStorage
	logger  *zap.SugaredLogger
}

func NewRestorer(service service.MetricsService, storage *storage.FileStorage, logger *zap.SugaredLogger) *Restorer {
	return &Restorer{
		service: service,
		storage: storage,
		logger:  logger,
	}
}

func (r *Restorer) Restore(ctx context.Context) error {
	restoreCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	gauges, counters, err := r.storage.Restore()
	if err != nil {
		var syntaxError *json.SyntaxError
		if errors.Is(err, io.EOF) {
			r.logger.Infoln("Snapshot file is empty, starting fresh")
			return nil
		}
		if errors.As(err, &syntaxError) {
			r.logger.Warn("Snapshot file is corrupted, starting fresh", zap.Error(err))
			return nil
		}
		r.logger.Fatalf("Failed to restore data: %v", err)
	}

	if len(gauges) == 0 && len(counters) == 0 {
		r.logger.Infoln("No metrics found in snapshot, starting fresh")
		return nil
	}

	if err := r.service.SaveAllMetrics(restoreCtx, gauges, counters); err != nil {
		r.logger.Errorf("Failed to load metrics into memory: %v", err)
		return err
	}

	r.logger.Infof("Restore completed. Loaded %d gauges, %d counters", len(gauges), len(counters))
	return nil
}
