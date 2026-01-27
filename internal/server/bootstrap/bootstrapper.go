package bootstrap

import (
	"context"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/mrPTqp/metrics/internal/server/backup"
	"github.com/mrPTqp/metrics/internal/server/config"
	"github.com/mrPTqp/metrics/internal/server/handler"
	"github.com/mrPTqp/metrics/internal/server/middleware"
	"github.com/mrPTqp/metrics/internal/server/repository"
	"github.com/mrPTqp/metrics/internal/server/service"
	"github.com/mrPTqp/metrics/internal/server/storage"
	"github.com/mrPTqp/metrics/internal/server/storage/migrations"
	"go.uber.org/zap"
)

type Bootstrapper struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewBootstrapper(cfg *config.Config, logger *zap.Logger) *Bootstrapper {
	return &Bootstrapper{cfg: cfg, logger: logger}
}

func (bs *Bootstrapper) MustRun(ctx context.Context) *AppComponents {
	bs.logger.Info("Starting application bootstrap...")

	var mr repository.MetricRepository
	if bs.cfg.DatabaseDsn != nil && *bs.cfg.DatabaseDsn != "" {
		bs.logger.Info("Applying database migrations...")
		if err := migrations.RunMigrations(*bs.cfg.DatabaseDsn, bs.logger); err != nil {
			bs.logger.Fatal("Migration failed", zap.Error(err))
		}
		bs.logger.Info("Migrations applied successfully or no changes")

		var err error
		mr, err = storage.NewPostgresStorage(*bs.cfg.DatabaseDsn, bs.logger)
		if err != nil {
			bs.logger.Fatal("Init postgres storage error", zap.Error(err))
		}

		checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		ok := mr.CheckStorageAvailability(checkCtx)
		cancel()
		if !ok {
			bs.logger.Fatal("Postgres connection check failed")
		}
	} else {
		mr = storage.NewMemStorage()
	}

	baseService := service.NewMetricsService(mr, bs.logger)
	var ms service.MetricsService = baseService

	p := storage.NewFileProducer(bs.cfg.File, bs.logger)
	c := storage.NewFileConsumer(bs.cfg.File, bs.logger)
	fsr := storage.NewFileStorage(p, c, mr, bs.cfg.SyncBackupToFile, bs.logger)

	var b *backup.Backuper
	if bs.cfg.SyncBackupToFile {
		ms = service.NewFileBackupService(baseService, fsr, bs.logger)
	} else {
		b = backup.NewBackuper(ms, fsr, bs.logger)
	}

	if bs.cfg.Restore {
		r := backup.NewRestorer(ms, fsr, bs.logger)
		restoreCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := r.Restore(restoreCtx)
		cancel()
		if err != nil {
			bs.logger.Error("Failed to restore metrics", zap.Error(err))
		}
	}

	mh := handler.NewMetricHandler(ms, bs.logger)

	mws := []func(http.Handler) http.Handler{
		middleware.LoggingMiddleware(bs.logger),
	}

	if bs.cfg.SecretKey != nil && *bs.cfg.SecretKey != "" {
		mws = append(mws, middleware.SignMiddleware(*bs.cfg.SecretKey))
	}

	mws = append(mws, middleware.GzipMiddleware)

	return &AppComponents{
		Config:      bs.cfg,
		Logger:      bs.logger,
		Repo:        mr,
		Backuper:    b,
		Handler:     mh,
		Middlewares: mws,
	}
}

type AppComponents struct {
	Config      *config.Config
	Logger      *zap.Logger
	Repo        repository.MetricRepository
	Handler     *handler.MetricHandler
	Backuper    *backup.Backuper
	Middlewares []func(http.Handler) http.Handler
}
