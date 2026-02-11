// internal/server/app/app.go
package app

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"github.com/mrPTqp/metrics/internal/server/backup"
	"github.com/mrPTqp/metrics/internal/server/bootstrap"
	"github.com/mrPTqp/metrics/internal/server/repository"
)

type App struct {
	cfg      *bootstrap.AppComponents
	server   *http.Server
	ticker   *time.Ticker
	shutdown sync.Once
	backuper *backup.Backuper
	repo     repository.MetricRepository
}

func NewApp(components *bootstrap.AppComponents) *App {
	r := chi.NewRouter()

	r.Get("/", wrap(components.Handler.CollectMetricsHandler, components.Middlewares...))
	if components.Config.DatabaseDsn != nil && *components.Config.DatabaseDsn != "" {
		r.Get("/ping", wrap(components.Handler.DBHealthCheckHandler, components.Middlewares...))
	}
	r.Route("/update", func(r chi.Router) {
		r.Post("/", wrap(components.Handler.SaveMetricHandlerJSON, components.Middlewares...))
		r.Post("/{type}/{name}/{value}", wrap(components.Handler.SaveMetricHandler, components.Middlewares...))
	})
	r.Route("/updates", func(r chi.Router) {
		r.Post("/", wrap(components.Handler.SaveMetricsHandlerJSON, components.Middlewares...))
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", wrap(components.Handler.ValueMetricHandlerJSON, components.Middlewares...))
		r.Get("/{type}/{name}", wrap(components.Handler.GetMetricHandler, components.Middlewares...))
	})
	r.Mount("/debug/pprof", http.DefaultServeMux)

	server := &http.Server{
		Addr:    components.Config.Address.String(),
		Handler: r,
	}

	return &App{
		cfg:      components,
		server:   server,
		ticker:   time.NewTicker(time.Duration(components.Config.StoreInterval) * time.Second),
		backuper: components.Backuper,
		repo:     components.Repo,
	}
}

func wrap(h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) http.HandlerFunc {
	var handler http.Handler = h
	for _, mw := range middlewares {
		handler = mw(handler)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	})
}

func (a *App) RunWithContext(ctx context.Context) {
	logger := a.cfg.Logger
	logger.Info("Starting HTTP server", zap.String("address", a.cfg.Config.Address.String()))

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server failed to start", zap.Error(err))
		}
	}()

	a.runBackgroundJobs(ctx)
}

func (a *App) runBackgroundJobs(ctx context.Context) {
	if a.backuper != nil {
		for {
			select {
			case <-ctx.Done():
				log := contextkey.LoggerFromContext(ctx)
				if log == nil {
					log = a.cfg.Logger // fallback
				}
				log.Info("Background job ticker stopped", zap.String("reason", ctx.Err().Error()))
				return
			case <-a.ticker.C:
				log := contextkey.LoggerFromContext(ctx)
				if log == nil {
					log = a.cfg.Logger
				}
				log.Debug("Triggering periodic backup...")
				backupCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				if err := a.backuper.Backup(backupCtx); err != nil {
					log.Error("Failed to backup metrics", zap.Error(err))
				}
				cancel()
			}
		}
	}
}

func (a *App) Shutdown(shutdownCtx context.Context) {
	a.shutdown.Do(func() {
		logger := a.cfg.Logger
		logger.Info("Shutting down server gracefully...")

		a.ticker.Stop()
		logger.Debug("Ticker stopped")

		if err := a.server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Server forced to shutdown", zap.Error(err))
		} else {
			logger.Info("Server stopped gracefully")
		}

		if a.cfg.EventBus != nil {
			logger.Info("Shutting down audit event bus...")
			if err := a.cfg.EventBus.ShutDown(shutdownCtx); err != nil {
				logger.Warn("EventBus shutdown timeout", zap.Error(err))
			} else {
				logger.Info("EventBus stopped")
			}
		}

		for _, proc := range a.cfg.AuditProcessors {
			if closer, ok := proc.(interface{ Shutdown(context.Context) error }); ok {
				if err := closer.Shutdown(shutdownCtx); err != nil {
					logger.Warn("Audit processor shutdown failed", zap.Error(err))
				}
			}
		}

		if a.backuper != nil {
			logger.Info("Saving metrics to file before shutdown...")
			backupCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
			err := a.backuper.Backup(backupCtx)
			cancel()
			if err != nil {
				logger.Warn("Failed to save metrics on shutdown", zap.Error(err))
			} else {
				logger.Info("Metrics saved on shutdown")
			}
		}

		if a.repo != nil {
			logger.Info("Closing storage connection...")
			if err := a.repo.Close(); err != nil {
				logger.Error("Error closing storage connection", zap.Error(err))
			} else {
				logger.Info("Storage connection closed")
			}
		}
	})
}
