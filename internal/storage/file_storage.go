// internal/storage/file_storage.go
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
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
	producer         *Producer
	consumer         *Consumer
	syncBackupToFile bool
}

func NewFileStorage(filePath string, syncBackupToFile bool, logger *zap.SugaredLogger) *FileStorage {
	producer, err := NewProducer(filePath)
	if err != nil {
		logger.Fatal("Failed to create file producer", zap.Error(err))
	}

	consumer, err := NewConsumer(filePath)
	if err != nil {
		producer.Close()
		logger.Fatal("Failed to create file consumer", zap.Error(err))
	}

	return &FileStorage{
		logger:           logger,
		filePath:         filePath,
		gauges:           make(map[string]float64),
		counters:         make(map[string]int64),
		producer:         producer,
		consumer:         consumer,
		syncBackupToFile: syncBackupToFile,
	}
}

func (fs *FileStorage) Restore() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	err := fs.consumer.Reopen()
	if err != nil {
		if os.IsNotExist(err) {
			fs.logger.Info("No snapshot file found, starting fresh")
			return nil
		}
		return err
	}
	defer fs.consumer.Close()

	var data snapshotData
	if err := fs.consumer.decoder.Decode(&data); err != nil {
		if errors.Is(err, io.EOF) {
			fs.logger.Warn("Snapshot file is empty, starting with empty state")
			return nil
		}
		var syntaxError *json.SyntaxError
		if errors.As(err, &syntaxError) {
			fs.logger.Warn("Snapshot file is corrupted, starting with empty state", zap.Error(err))
			return nil
		}
		return fmt.Errorf("failed to decode snapshot: %w", err)
	}

	fs.gauges = data.Gauges
	fs.counters = data.Counters
	return nil
}

func (fs *FileStorage) Backup() error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if err := fs.producer.Close(); err != nil {
		fs.logger.Warn("Error closing producer before backup", zap.Error(err))
	}

	newProducer, err := NewProducer(fs.filePath)
	if err != nil {
		fs.logger.Error("Failed to create producer for backup", zap.Error(err))
		return err
	}
	defer newProducer.Close()

	data := snapshotData{
		Gauges:   fs.gauges,
		Counters: fs.counters,
	}

	if err := newProducer.encoder.Encode(&data); err != nil {
		fs.logger.Error("Failed to encode snapshot", zap.Error(err))
		return err
	}

	return nil
}

func (fs *FileStorage) SaveGauge(key string, value *float64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.gauges[key] = *value

	if fs.syncBackupToFile {
		if err := fs.Backup(); err != nil {
			fs.logger.Error("Failed to backup gauge", zap.Error(err))
		}
	}
	return nil
}

func (fs *FileStorage) SaveCounter(key string, value *int64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.counters[key] += *value

	if fs.syncBackupToFile {
		if err := fs.Backup(); err != nil {
			fs.logger.Error("Failed to backup counter", zap.Error(err))
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

func (fs *FileStorage) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	var err1, err2 error
	if fs.producer != nil {
		err1 = fs.producer.Close()
	}
	if fs.consumer != nil {
		err2 = fs.consumer.Close()
	}
	if err1 != nil {
		return err1
	}
	return err2
}

func (fs *FileStorage) SaveAllMetrics(gauges map[string]float64, counters map[string]int64) error {
	for key, value := range gauges {
		if err := fs.SaveGauge(key, &value); err != nil {
			return err
		}
	}
	for key, value := range counters {
		if err := fs.SaveCounter(key, &value); err != nil {
			return err
		}
	}

	return nil
}

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(filePath string) (*Producer, error) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}
	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file     *os.File
	decoder  *json.Decoder
	filePath string
}

func NewConsumer(filePath string) (*Consumer, error) {
	return &Consumer{filePath: filePath}, nil
}

func (c *Consumer) Reopen() error {
	if c.file != nil {
		_ = c.file.Close()
	}
	file, err := os.Open(c.filePath)
	if err != nil {
		return err
	}
	c.file = file
	c.decoder = json.NewDecoder(file)
	return nil
}

func (c *Consumer) Close() error {
	if c.file == nil {
		return nil
	}
	return c.file.Close()
}
