package bootstrap

import (
	"database/sql"
	"net/http"

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
	logger *zap.SugaredLogger
	db     *sql.DB
}

func NewBootstrapper(cfg *config.Config, logger *zap.SugaredLogger) *Bootstrapper {
	return &Bootstrapper{cfg: cfg, logger: logger}
}

func (bs *Bootstrapper) MustRun() *AppComponents {
	bs.logger.Info("Starting application bootstrap...")

	var mr repository.MetricRepository
	if bs.cfg.DatabaseDsn != nil && *bs.cfg.DatabaseDsn != "" {
		var err error
		bs.logger.Info("Applying database migrations...")
		if err = migrations.RunMigrations(*bs.cfg.DatabaseDsn, bs.logger); err != nil {
			bs.logger.Panicf("Migration failed: %v", err)
		}
		bs.logger.Info("Migrations applied successfully or no changes")

		mr, err = storage.NewPostgresStorage(*bs.cfg.DatabaseDsn, bs.logger)
		if err != nil {
			bs.logger.Panic("init postgres error", err)
		}

		if !mr.CheckStorageAvailability() {
			bs.logger.Panic("postgres connection error")
		}
	} else {
		mr = storage.NewMemStorage(bs.logger)
	}

	var ms service.MetricsService
	baseService := service.NewMetricsService(mr, bs.logger)
	ms = baseService

	p := storage.NewFileProducer(bs.cfg.File, bs.logger)
	c := storage.NewFileConsumer(bs.cfg.File, bs.logger)
	fsr := storage.NewFileStorage(p, c, mr, bs.cfg.SyncBackupToFile, bs.logger)

	var b *backup.Backuper
	if bs.cfg.SyncBackupToFile {
		ms = service.NewFileBackupService(baseService, fsr, bs.logger)
	} else {
		b = backup.NewBackuper(ms, fsr, bs.logger)		
	}

	var r *backup.Restorer
	if bs.cfg.Restore {
		r = backup.NewRestorer(ms, fsr, bs.logger)
		r.Restore()
	}

	mh := handler.NewMetricHandler(ms, bs.logger)
	mws := []func(h http.HandlerFunc, logger *zap.SugaredLogger) http.HandlerFunc{
		middleware.LoggingMiddleware,
	}

	if bs.cfg.SecretKey != nil && *bs.cfg.SecretKey != "" {
		mws = append(mws, func(h http.HandlerFunc, logger *zap.SugaredLogger) http.HandlerFunc {
			return middleware.SignMiddleware(h, *bs.cfg.SecretKey, bs.logger)
		})
	}

	mws = append(mws, middleware.GzipMiddleware)

	return &AppComponents{
		Config:   bs.cfg,
		Logger:   bs.logger,
		DB:       bs.db,
		Repo:     &mr,
		Backuper: b,
		Handler:     mh,
		Middlewares: mws,
	}
}

type AppComponents struct {
	Config      *config.Config
	Logger      *zap.SugaredLogger
	DB          *sql.DB
	Repo        *repository.MetricRepository
	Handler     *handler.MetricHandler
	Backuper    *backup.Backuper
	Middlewares []func(h http.HandlerFunc, logger *zap.SugaredLogger) http.HandlerFunc
}
