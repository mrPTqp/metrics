// internal/storage/file_storage.go
package storage

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"go.uber.org/zap"
)

type snapshotData struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

type FileStorage struct {
	mu               sync.RWMutex
	logger           *zap.SugaredLogger
	filePath         string
	gauges           map[string]float64
	counters         map[string]int64
	syncBackupToFile bool
}

func NewFileStorage(filePath string, syncBackupToFile bool, logger *zap.SugaredLogger) *FileStorage {
	return &FileStorage{
		logger:           logger,
		filePath:         filePath,
		gauges:           make(map[string]float64),
		counters:         make(map[string]int64),
		syncBackupToFile: syncBackupToFile,
	}
}

func (fs *FileStorage) Backup() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data := snapshotData{
		Gauges:   fs.gauges,
		Counters: fs.counters,
	}

	file, err := os.Create(fs.filePath)
	if err != nil {
		fs.logger.Error("Failed to create backup file", zap.Error(err))
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(data); err != nil {
		fs.logger.Error("Failed to encode data", zap.Error(err))
		return err
	}

	if err := file.Sync(); err != nil {
		fs.logger.Error("Failed to sync file", zap.Error(err))
		return err
	}

	fs.logger.Info("Backup completed", zap.String("file", fs.filePath))
	return nil
}

func (fs *FileStorage) Restore() (map[string]float64, map[string]int64, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data := snapshotData{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}

	file, err := os.Open(fs.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fs.logger.Info("No snapshot file, starting fresh")
			return data.Gauges, data.Counters, nil
		}
		return nil, nil, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		if errors.Is(err, io.EOF) || errors.As(err, new(*json.SyntaxError)) {
			fs.logger.Warn("Corrupted or empty snapshot, starting fresh")
			return make(map[string]float64), make(map[string]int64), nil
		}
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
		err := fs.Backup()
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
		err := fs.Backup()
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
	fs.gauges = copyMap(gauges)
	fs.counters = copyMapInt64(counters)

	return fs.Backup()
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
