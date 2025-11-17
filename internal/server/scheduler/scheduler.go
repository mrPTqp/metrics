package scheduler

import (
	"time"

	"github.com/mrPTqp/metrics/internal/server/backup"

	"go.uber.org/zap"
)

type FileBackupScheduler struct {
	backuper *backup.Backuper
	logger  *zap.SugaredLogger
}

func NewScheduler(backuper *backup.Backuper, logger *zap.SugaredLogger) *FileBackupScheduler {
	return &FileBackupScheduler{
		backuper: backuper,
		logger:  logger,
	}
}

func (fbs *FileBackupScheduler) Start(storeInterval int, fileStoragePath string) {
	ticker := time.NewTicker(time.Duration(storeInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fbs.backuper.Backup()
	}
}
