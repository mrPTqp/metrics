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

type FileProducer struct {
	filePath string
	logger   *zap.SugaredLogger
	mu       sync.Mutex
}

func NewFileProducer(filePath string, logger *zap.SugaredLogger) *FileProducer {
	return &FileProducer{
		filePath: filePath,
		logger:   logger,
	}
}

func (fp *FileProducer) WriteData(data snapshotData) error {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	file, err := os.Create(fp.filePath)
	if err != nil {
		fp.logger.Error("Failed to create backup file", zap.Error(err))
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(data); err != nil {
		fp.logger.Error("Failed to encode data", zap.Error(err))
		return err
	}

	if err := file.Sync(); err != nil {
		fp.logger.Error("Failed to sync file", zap.Error(err))
		return err
	}

	fp.logger.Info("Backup completed", zap.String("file", fp.filePath))
	return nil
}

type FileConsumer struct {
	filePath string
	logger   *zap.SugaredLogger
	mu       sync.Mutex
}

func NewFileConsumer(filePath string, logger *zap.SugaredLogger) *FileConsumer {
	return &FileConsumer{
		filePath: filePath,
		logger:   logger,
	}
}

func (fc *FileConsumer) ReadData() (snapshotData, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	data := snapshotData{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}

	file, err := os.Open(fc.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fc.logger.Info("No snapshot file, starting fresh")
			return data, nil
		}
		return data, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		if errors.Is(err, io.EOF) || errors.As(err, new(*json.SyntaxError)) {
			fc.logger.Warn("Corrupted or empty snapshot, starting fresh")
			return snapshotData{
				Gauges:   make(map[string]float64),
				Counters: make(map[string]int64),
			}, nil
		}
		return data, err
	}

	return data, nil
}

func (fc *FileConsumer) CheckFileAccess() bool {
	_, err := os.Stat(fc.filePath)
	if err != nil {
		if os.IsNotExist(err) || os.IsPermission(err) {
			fc.logger.Info("File is not available", zap.String("file", fc.filePath), zap.Error(err))
			return false
		}
		fc.logger.Info("Error accessing file", zap.String("file", fc.filePath), zap.Error(err))
		return false
	}
	return true
}
