package metrics

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
	cancel   context.CancelFunc
	shutdown sync.Once
	backuper *backup.Backuper
	repo     *repository.MetricRepository
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
		cfg:    components,
		server: server,
		logger: components.Logger,
		ticker: time.NewTicker(time.Duration(components.Config.StoreInterval)),
	}
}

func wrap(h http.HandlerFunc, logger *zap.SugaredLogger, mwFuncs ...func(http.HandlerFunc, *zap.SugaredLogger) http.HandlerFunc) http.HandlerFunc {
	for i := len(mwFuncs) - 1; i >= 0; i-- {
		h = mwFuncs[i](h, logger)
	}
	return h
}

func (a *App) Run() {
	a.logger.Info("Starting HTTP server", zap.String("address", a.cfg.Config.Address.String()))

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("HTTP server failed to start", zap.Error(err))
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	go a.runBackgroundJobs(ctx)
}

func (a *App) runBackgroundJobs(ctx context.Context) {
	if a.backuper != nil {
		for {
			select {
			case <-ctx.Done():
				a.logger.Info("Background job ticker stopped", zap.String("reason", ctx.Err().Error()))
				return
			case <-a.ticker.C:
				a.logger.Debug("backup metrics")
				a.backuper.Backup()
			}
		}
	}
}

func (a *App) Shutdown(ctx context.Context) {
	a.shutdown.Do(func() {
		a.logger.Info("Shutting down server gracefully...")

		if a.cancel != nil {
			a.cancel()
		}

		a.ticker.Stop()
		a.logger.Debug("Ticker stopped")

		if err := a.server.Shutdown(ctx); err != nil {
			a.logger.Errorf("Server forced to shutdown: %v", err)
		} else {
			a.logger.Info("Server stopped gracefully")
		}

		if a.backuper != nil {
			a.logger.Info("Saving metrics to file before shutdown...")
			a.backuper.Backup()
		}

		if a.repo != nil {
			a.logger.Info("Closing storage connection...")
			repo := *a.repo
			if err := repo.Close(); err != nil {
				a.logger.Errorf("Error closing storage connection: %v", err)
			} else {
				a.logger.Info("storage connection closed")
			}
		}
	})
}
