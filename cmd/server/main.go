package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/logger"
	"github.com/mrPTqp/metrics/internal/server/backup"
	"github.com/mrPTqp/metrics/internal/server/config"
	"github.com/mrPTqp/metrics/internal/server/handler"
	"github.com/mrPTqp/metrics/internal/server/metrics"
	"github.com/mrPTqp/metrics/internal/server/middleware"
	"github.com/mrPTqp/metrics/internal/server/repository"
	"github.com/mrPTqp/metrics/internal/server/scheduler"
	"github.com/mrPTqp/metrics/internal/server/service"
	"github.com/mrPTqp/metrics/internal/server/storage"
	"github.com/mrPTqp/metrics/internal/server/storage/migrations"
)

func main() {
	sugar := logger.NewSugarLogger()

	cfg := config.LoadConfig()
	sugar.Infow("configuration created", "config", cfg)

	var mr repository.MetricRepository
	if cfg.DatabaseDsn != nil && *cfg.DatabaseDsn != "" {
		var err error
		sugar.Info("Applying database migrations...")
		if err = migrations.RunMigrations(*cfg.DatabaseDsn, sugar); err != nil {
			sugar.Panicf("Migration failed: %v", err)
		}
		sugar.Info("Migrations applied successfully or no changes")

		mr, err = storage.NewPostgresStorage(*cfg.DatabaseDsn, sugar)
		if err != nil {
			sugar.Panic("init postgres error", err)
		}

		if !mr.CheckStorageAvailability() {
			sugar.Panic("postgres connection error")
		}
	} else {
		mr = storage.NewMemStorage(sugar)
	}

	var ms service.MetricsService
	baseService := service.NewMetricsService(mr, sugar)
	ms = baseService

	p := storage.NewFileProducer(cfg.File, sugar)
	c := storage.NewFileConsumer(cfg.File, sugar)
	fsr := storage.NewFileStorage(p, c, mr, cfg.SyncBackupToFile, sugar)

	var b *backup.Backuper
	var sc *scheduler.FileBackupScheduler
	if cfg.SyncBackupToFile {
		ms = service.NewFileBackupService(baseService, fsr, sugar)
	} else {
		b = backup.NewBackuper(ms, fsr, sugar)
		sc = scheduler.NewScheduler(b, sugar)
		go sc.Start(cfg.StoreInterval, cfg.File)
	}

	var r *backup.Restorer
	if cfg.Restore {
		r = backup.NewRestorer(ms, fsr, sugar)
		r.Restore()
	}

	mh := handler.NewMetricHandler(ms, sugar)
	mws := []func(h http.HandlerFunc, sugar *zap.SugaredLogger) http.HandlerFunc{
		middleware.LoggingMiddleware,
		middleware.GzipMiddleware,
	}

	if cfg.SecretKey != nil && *cfg.SecretKey != "" {
		mws = append(mws, func(h http.HandlerFunc, sugar *zap.SugaredLogger) http.HandlerFunc {
			return middleware.SignMiddleware(h, *cfg.SecretKey, sugar)
		})
	}

	srv := metrics.StartMetricsServer(mh, mws, cfg, sugar)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	sugar.Info("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		sugar.Errorf("Server forced to shutdown: %v", err)
	} else {
		sugar.Info("Server stopped gracefully")
	}

	if b != nil {
		sugar.Info("Saving metrics to file before shutdown...")
		b.Backup()
	}

	if mr != nil {
		sugar.Info("Closing storage connection...")
		if err := mr.Close(); err != nil {
			sugar.Errorf("Error closing storage connection: %v", err)
		} else {
			sugar.Info("storage connection closed")
		}
	}
}
