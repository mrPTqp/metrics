// internal/server/app/app.go
package app

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/server/backup"
	"github.com/mrPTqp/metrics/internal/server/bootstrap"
	"github.com/mrPTqp/metrics/internal/server/repository"
)

type App struct {
	cfg      *bootstrap.AppComponents
	server   *http.Server
	logger   *zap.SugaredLogger
	ticker   *time.Ticker
	shutdown sync.Once
	backuper *backup.Backuper
	repo     repository.MetricRepository
}

func NewApp(components *bootstrap.AppComponents) *App {
	r := chi.NewRouter()
	r.Get("/", wrap(components.Handler.CollectMetricsHandler, components.Logger, components.Middlewares...))
	if components.Config.DatabaseDsn != nil && *components.Config.DatabaseDsn != "" {
		r.Get("/ping", wrap(components.Handler.DBHealthCheckHandler, components.Logger, components.Middlewares...))
	}
	r.Route("/update", func(r chi.Router) {
		r.Post("/", wrap(components.Handler.SaveMetricHandlerJSON, components.Logger, components.Middlewares...))
		r.Post("/{type}/{name}/{value}", wrap(components.Handler.SaveMetricHandler, components.Logger, components.Middlewares...))
	})
	r.Route("/updates", func(r chi.Router) {
		r.Post("/", wrap(components.Handler.SaveMetricsHandlerJSON, components.Logger, components.Middlewares...))
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", wrap(components.Handler.ValueMetricHandlerJSON, components.Logger, components.Middlewares...))
		r.Get("/{type}/{name}", wrap(components.Handler.GetMetricHandler, components.Logger, components.Middlewares...))
	})

	server := &http.Server{
		Addr:    components.Config.Address.String(),
		Handler: r,
	}

	return &App{
		cfg:      components,
		server:   server,
		logger:   components.Logger,
		ticker:   time.NewTicker(time.Duration(components.Config.StoreInterval) * time.Second),
		backuper: components.Backuper,
		repo:     components.Repo,
	}
}

func wrap(h http.HandlerFunc, logger *zap.SugaredLogger, mwFuncs ...func(http.HandlerFunc, *zap.SugaredLogger) http.HandlerFunc) http.HandlerFunc {
	for i := len(mwFuncs) - 1; i >= 0; i-- {
		h = mwFuncs[i](h, logger)
	}
	return h
}

func (a *App) RunWithContext(ctx context.Context) {
	a.logger.Infow("Starting HTTP server", zap.String("address", a.cfg.Config.Address.String()))

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("HTTP server failed to start", zap.Error(err))
		}
	}()

	a.runBackgroundJobs(ctx)
}

func (a *App) runBackgroundJobs(ctx context.Context) {
	if a.backuper != nil {
		for {
			select {
			case <-ctx.Done():
				a.logger.Debug("Background job ticker stopped due to context cancellation")
				return
			case <-a.ticker.C:
				a.logger.Debug("Triggering periodic backup...")
				backupCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				if err := a.backuper.Backup(backupCtx); err != nil {
					a.logger.Errorf("Failed to backup metrics: %v", err)
				}
				cancel()
			}
		}
	}
}

func (a *App) Shutdown(shutdownCtx context.Context) {
	a.shutdown.Do(func() {
		a.logger.Infoln("Shutting down server gracefully...")

		a.ticker.Stop()
		a.logger.Debug("Ticker stopped")

		if err := a.server.Shutdown(shutdownCtx); err != nil {
			a.logger.Errorf("Server forced to shutdown: %v", err)
		} else {
			a.logger.Infoln("Server stopped gracefully")
		}

		if a.backuper != nil {
			a.logger.Infoln("Saving metrics to file before shutdown...")
			backupCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
			err := a.backuper.Backup(backupCtx)
			cancel()
			if err != nil {
				a.logger.Warn("Failed to save metrics on shutdown")
			} else {
				a.logger.Infoln("Metrics saved on shutdown")
			}
		}

		if a.repo != nil {
			a.logger.Infoln("Closing storage connection...")
			if err := a.repo.Close(); err != nil {
				a.logger.Errorf("Error closing storage connection: %v", err)
			} else {
				a.logger.Infoln("Storage connection closed")
			}
		}
	})
}
