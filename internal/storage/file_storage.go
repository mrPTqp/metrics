package storage

import (
	"errors"
	"sync"

	"go.uber.org/zap"
)

type FileStorage struct {
	producer         *FileProducer
	consumer         *FileConsumer
	gauges           map[string]float64
	counters         map[string]int64
	syncBackupToFile bool
	mu               sync.RWMutex
	logger           *zap.SugaredLogger
}

func NewFileStorage(producer *FileProducer, consumer *FileConsumer, syncBackupToFile bool, logger *zap.SugaredLogger) *FileStorage {
	return &FileStorage{

		producer:         producer,
		consumer:         consumer,
		gauges:           make(map[string]float64),
		counters:         make(map[string]int64),
		syncBackupToFile: syncBackupToFile,
		logger:           logger,
	}
}

func (fs *FileStorage) Backup() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data := snapshotData{
		Gauges:   fs.gauges,
		Counters: fs.counters,
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

	fs.gauges = data.Gauges
	fs.counters = data.Counters

	return fs.gauges, fs.counters, nil
}

func (fs *FileStorage) SaveGauge(key string, value *float64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.gauges[key] = *value

	if fs.syncBackupToFile {
		data := snapshotData{
			Gauges:   fs.gauges,
			Counters: fs.counters,
		}
		err := fs.producer.WriteData(data)
		if err != nil {
			fs.logger.Errorf("failed backup gauge metric %s", key)
		}
	}
	return nil
}

func (fs *FileStorage) SaveCounter(key string, value *int64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.counters[key] += *value

	if fs.syncBackupToFile {
		data := snapshotData{
			Gauges:   fs.gauges,
			Counters: fs.counters,
		}
		err := fs.producer.WriteData(data)
		if err != nil {
			fs.logger.Errorf("failed backup counter metric %s", key)
		}
	}
	return nil
}

func (fs *FileStorage) GetGauge(key string) (float64, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	if val, ok := fs.gauges[key]; ok {
		return val, nil
	}
	return 0, errors.New("gauge not found")
}

func (fs *FileStorage) GetCounter(key string) (int64, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	if val, ok := fs.counters[key]; ok {
		return val, nil
	}
	return 0, errors.New("counter not found")
}

func (fs *FileStorage) ListGauges() (map[string]float64, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	copy := make(map[string]float64, len(fs.gauges))
	for k, v := range fs.gauges {
		copy[k] = v
	}
	return copy, nil
}

func (fs *FileStorage) ListCounters() (map[string]int64, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	copy := make(map[string]int64, len(fs.counters))
	for k, v := range fs.counters {
		copy[k] = v
	}
	return copy, nil
}

func (fs *FileStorage) SaveAllMetrics(gauges map[string]float64, counters map[string]int64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.gauges = copyMap(gauges)
	fs.counters = copyMapInt64(counters)

	data := snapshotData{
		Gauges:   fs.gauges,
		Counters: fs.counters,
	}

	return fs.producer.WriteData(data)
}

func copyMap(src map[string]float64) map[string]float64 {
	dst := make(map[string]float64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func copyMapInt64(src map[string]int64) map[string]int64 {
	dst := make(map[string]int64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (fs *FileStorage) CheckStorageAvailability() bool {
	return fs.consumer.CheckFileAccess()
}
