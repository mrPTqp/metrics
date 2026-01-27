package storage

import (
	"context"
	"errors"
	"sync"

	"go.uber.org/zap"
)

type MetricsProvider interface {
	ListGauges(context.Context) (map[string]float64, error)
	ListCounters(context.Context) (map[string]int64, error)
}

type FileStorage struct {
	producer         *FileProducer
	consumer         *FileConsumer
	metricsProvider  MetricsProvider
	syncBackupToFile bool
	mu               sync.RWMutex
	logger           *zap.Logger
}

func NewFileStorage(producer *FileProducer, consumer *FileConsumer, metricsProvider MetricsProvider, syncBackupToFile bool, logger *zap.Logger) *FileStorage {
	return &FileStorage{
		producer:         producer,
		consumer:         consumer,
		metricsProvider:  metricsProvider,
		syncBackupToFile: syncBackupToFile,
		logger:           logger,
	}
}

func (fs *FileStorage) Backup() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	gauges, err := fs.metricsProvider.ListGauges(context.Background())
	if err != nil {
		fs.logger.Error("Failed to list gauges for backup", zap.Error(err))
		return err
	}

	counters, err := fs.metricsProvider.ListCounters(context.Background())
	if err != nil {
		fs.logger.Error("Failed to list counters for backup", zap.Error(err))
		return err
	}

	data := snapshotData{
		Gauges:   gauges,
		Counters: counters,
	}

	return fs.producer.WriteData(data)
}

func (fs *FileStorage) Restore() (map[string]float64, map[string]int64, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := fs.consumer.ReadData()
	if err != nil {
		return nil, nil, err
	}

	return data.Gauges, data.Counters, nil
}

func (fs *FileStorage) backupAllMetrics(metricType, metricName string) {
	if !fs.syncBackupToFile {
		return
	}

	gauges, err := fs.metricsProvider.ListGauges(context.Background())
	if err != nil {
		fs.logger.Error("Failed to get gauges for backup", zap.Error(err))
		return
	}

	counters, err := fs.metricsProvider.ListCounters(context.Background())
	if err != nil {
		fs.logger.Error("Failed to get counters for backup", zap.Error(err))
		return
	}

	data := snapshotData{
		Gauges:   gauges,
		Counters: counters,
	}

	if err := fs.producer.WriteData(data); err != nil {
		fs.logger.Error("Failed to backup metrics",
			zap.String("type", metricType),
			zap.String("name", metricName),
			zap.Error(err))
	}
}

func (fs *FileStorage) SaveGauge(ctx context.Context, key string, value *float64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.backupAllMetrics("gauge", key)
	return nil
}

func (fs *FileStorage) SaveCounter(ctx context.Context, key string, value *int64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.backupAllMetrics("counter", key)
	return nil
}

func (fs *FileStorage) GetGauge(ctx context.Context, key string) (float64, error) {
	return 0, errors.New("FileStorage does not store data, use the main storage")
}

func (fs *FileStorage) GetCounter(ctx context.Context, key string) (int64, error) {
	return 0, errors.New("FileStorage does not store data, use the main storage")
}

func (fs *FileStorage) ListGauges(ctx context.Context) (map[string]float64, error) {
	return nil, errors.New("FileStorage does not store data, use the main storage")
}

func (fs *FileStorage) ListCounters(ctx context.Context) (map[string]int64, error) {
	return nil, errors.New("FileStorage does not store data, use the main storage")
}

func (fs *FileStorage) SaveAllMetrics(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data := snapshotData{
		Gauges:   gauges,
		Counters: counters,
	}

	err := fs.producer.WriteData(data)
	if err != nil {
		fs.logger.Error("Failed to save all metrics to file", zap.Error(err))
	}
	return err
}

func (fs *FileStorage) CheckStorageAvailability(ctx context.Context) bool {
	return fs.consumer.CheckFileAccess()
}

func (fs *FileStorage) Close() error {
	fs.logger.Info("FileStorage closed")
	return nil
}
