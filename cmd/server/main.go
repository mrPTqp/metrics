package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/backup"
	"github.com/mrPTqp/metrics/internal/handler"
	"github.com/mrPTqp/metrics/internal/logger"
	"github.com/mrPTqp/metrics/internal/middleware"
	"github.com/mrPTqp/metrics/internal/repository"
	"github.com/mrPTqp/metrics/internal/scheduler"
	"github.com/mrPTqp/metrics/internal/service"
	"github.com/mrPTqp/metrics/internal/storage"
)

func main() {
	sugar := logger.NewSugarLogger()

	cfg := LoadConfig()
	sugar.Infow("configuration loaded", "config", cfg)

	var mr repository.MetricRepository
	if cfg.DatabaseDsn != "" {
		var err error

		sugar.Info("Applying database migrations...")
		if err = runMigrations(cfg.DatabaseDsn, sugar); err != nil {
			sugar.Panicf("Failed to run migrations: %v", err)
		}
		sugar.Info("Migrations applied successfully or no changes")

		mr, err = storage.NewPostgresStorage(cfg.DatabaseDsn, sugar)
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
	srv := startMetricsServer(mh, mws, cfg, sugar)

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

func runMigrations(dsn string, sugar *zap.SugaredLogger) error {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	current, _, _ := m.Version()
	sugar.Infof("Current migration version: %d", current)

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			sugar.Info("No migrations to apply")
			return nil
		}

		sugar.Errorf("Migration error: %v. Attempting rollback...", err)
		if rollbackErr := m.Down(); rollbackErr != nil {
			sugar.Errorf("Rollback after migration failure failed: %v", rollbackErr)
		} else {
			sugar.Info("Rollback after migration failure succeeded")
		}

		return err
	}

	newVersion, _, _ := m.Version()
	if newVersion > current {
		sugar.Infof("Successfully migrated to version %d", newVersion)
	} else {
		sugar.Info("No new migrations found")
	}

	return nil
}

func startMetricsServer(mh *handler.MetricHandler, mws []func(h http.HandlerFunc, sugar *zap.SugaredLogger) http.HandlerFunc, cfg *Config, sugar *zap.SugaredLogger) *http.Server {
	r := chi.NewRouter()
	r.Get("/", wrap(mh.CollectMetricsHandler, sugar, mws...))
	if cfg.DatabaseDsn != "" {
		r.Get("/ping", wrap(mh.DBHealthCheckHandler, sugar, mws...))
	}
	r.Route("/update", func(r chi.Router) {
		r.Post("/", wrap(mh.SaveMetricHandlerJSON, sugar, mws...))
		r.Post("/{type}/{name}/{value}", wrap(mh.SaveMetricHandler, sugar, mws...))
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", wrap(mh.ValueMetricHandlerJSON, sugar, mws...))
		r.Get("/{type}/{name}", wrap(mh.GetMetricHandler, sugar, mws...))
	})

	srv := &http.Server{
		Addr:    cfg.Address.String(),
		Handler: r,
	}

	go func() {
		sugar.Infof("Server is running on %s", cfg.Address.String())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalf("Server failed: %v", err)
		}
	}()

	return srv
}

func wrap(h http.HandlerFunc, logger *zap.SugaredLogger, mwFuncs ...func(http.HandlerFunc, *zap.SugaredLogger) http.HandlerFunc) http.HandlerFunc {
	for i := len(mwFuncs) - 1; i >= 0; i-- {
		h = mwFuncs[i](h, logger)
	}
	return h
}
