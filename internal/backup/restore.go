package backup

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/mrPTqp/metrics/internal/service"
	"github.com/mrPTqp/metrics/internal/storage"
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

func (r *Restorer) Restore() {
	gauges, counters, err := r.storage.Restore()
	if err != nil {
		var syntaxError *json.SyntaxError
		if errors.Is(err, io.EOF) {
			r.logger.Info("Snapshot file is empty, starting fresh")
			return
		}
		if errors.As(err, &syntaxError) {
			r.logger.Warn("Snapshot file is corrupted, starting fresh", zap.Error(err))
			return
		}
		r.logger.Fatalf("Failed to restore data: %v", err)
	}

	if err := r.service.SaveAllMetrics(gauges, counters); err != nil {
		r.logger.Errorf("Failed to load metrics into memory: %v", err)
	}

	r.logger.Infof("Restore completed. Loaded %d gauges, %d counters", len(gauges), len(counters))
}
