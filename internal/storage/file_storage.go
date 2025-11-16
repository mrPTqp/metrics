package storage

import (
	"errors"
	"sync"

	"go.uber.org/zap"
)

type MetricsProvider interface {
	ListGauges() (map[string]float64, error)
	ListCounters() (map[string]int64, error)
}

type FileStorage struct {
	producer         *FileProducer
	consumer         *FileConsumer
	metricsProvider  MetricsProvider
	syncBackupToFile bool
	mu               sync.RWMutex
	logger           *zap.SugaredLogger
}

func NewFileStorage(producer *FileProducer, consumer *FileConsumer, metricsProvider MetricsProvider, syncBackupToFile bool, logger *zap.SugaredLogger) *FileStorage {
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

	gauges, err := fs.metricsProvider.ListGauges()
	if err != nil {
		return err
	}

	counters, err := fs.metricsProvider.ListCounters()
	if err != nil {
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

	gauges, err := fs.metricsProvider.ListGauges()
	if err != nil {
		fs.logger.Errorf("failed to get gauges for backup: %v", err)
		return
	}

	counters, err := fs.metricsProvider.ListCounters()
	if err != nil {
		fs.logger.Errorf("failed to get counters for backup: %v", err)
		return
	}

	data := snapshotData{
		Gauges:   gauges,
		Counters: counters,
	}

	if err := fs.producer.WriteData(data); err != nil {
		fs.logger.Errorf("failed backup %s metric %s: %v", metricType, metricName, err)
	}
}

func (fs *FileStorage) SaveGauge(key string, value *float64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.backupAllMetrics("gauge", key)
	return nil
}

func (fs *FileStorage) SaveCounter(key string, value *int64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.backupAllMetrics("counter", key)
	return nil
}

func (fs *FileStorage) GetGauge(key string) (float64, error) {
	return 0, errors.New("FileStorage does not store data, use the main storage")
}

func (fs *FileStorage) GetCounter(key string) (int64, error) {
	return 0, errors.New("FileStorage does not store data, use the main storage")
}

func (fs *FileStorage) ListGauges() (map[string]float64, error) {
	return nil, errors.New("FileStorage does not store data, use the main storage")
}

func (fs *FileStorage) ListCounters() (map[string]int64, error) {
	return nil, errors.New("FileStorage does not store data, use the main storage")
}

func (fs *FileStorage) SaveAllMetrics(gauges map[string]float64, counters map[string]int64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data := snapshotData{
		Gauges:   gauges,
		Counters: counters,
	}

	return fs.producer.WriteData(data)
}

func (fs *FileStorage) CheckStorageAvailability() bool {
	return fs.consumer.CheckFileAccess()
}
