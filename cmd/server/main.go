package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/handler"
	"github.com/mrPTqp/metrics/internal/middleware"
	"github.com/mrPTqp/metrics/internal/scheduler"
	"github.com/mrPTqp/metrics/internal/service"
	"github.com/mrPTqp/metrics/internal/storage"
)

func main() {
	var sugar *zap.SugaredLogger
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	sugar = logger.Sugar()

	cfg := LoadConfig()
	sugar.Infow("Configuration loaded", "config", cfg)

	msr := storage.NewMemStorage(sugar)
	fsr := storage.NewFileStorage(cfg.File, cfg.SyncBackupToFile, sugar)
	ms := service.NewMetricsService(msr, fsr, sugar)
	mh := handler.NewMetricHandler(ms, sugar)

	if cfg.Restore {
		restore(fsr, sugar, ms)
	}

	mws := []func(h http.HandlerFunc, sugar *zap.SugaredLogger) http.HandlerFunc{
		middleware.LoggingMiddleware,
		middleware.GzipMiddleware,
	}
	srv := startMetricsServer(mh, mws, cfg, sugar)

	var sc *scheduler.FileBackupScheduler
	if !cfg.SyncBackupToFile {
		sc = scheduler.NewScheduler(ms, sugar)
		go sc.Start(cfg.StoreInterval, cfg.File)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	sugar.Info("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		sugar.Errorf("Server forced to shutdown: %v", err)
	} else {
		sugar.Info("Server stopped")
	}

	if sc != nil {
		sugar.Info("Saving metrics to file before shutdown...")
		sc.Backup()
	}
}

func restore(fsr *storage.FileStorage, sugar *zap.SugaredLogger, ms *service.BaseMetricService) {
	gauges, counters, err := fsr.Restore()
	if err != nil {
		var syntaxError *json.SyntaxError
		if errors.Is(err, io.EOF) {
			sugar.Info("Snapshot file is empty, starting fresh")
			return
		}
		if errors.As(err, &syntaxError) {
			sugar.Warn("Snapshot file is corrupted, starting fresh", zap.Error(err))
			return
		}
		sugar.Fatalf("Failed to restore data: %v", err)
	}

	if err := ms.SaveAllMetrics(gauges, counters); err != nil {
		sugar.Errorf("Failed to load metrics into memory: %v", err)
	}

	sugar.Infof("Restore completed. Loaded %d gauges, %d counters", len(gauges), len(counters))
}

func startMetricsServer(mh *handler.MetricHandler, mws []func(h http.HandlerFunc, sugar *zap.SugaredLogger) http.HandlerFunc, cfg *Config, sugar *zap.SugaredLogger) *http.Server {
	r := chi.NewRouter()
	r.Get("/", wrap(
		mh.CollectMetricsHandler,
		sugar,
		mws...,
	))
	r.Route("/update", func(r chi.Router) {
		r.Post("/", wrap(
			mh.SaveMetricHandlerJSON,
			sugar,
			mws...,
		))
		r.Post("/{type}/{name}/{value}", wrap(
			mh.SaveMetricHandler,
			sugar,
			mws...,
		))
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", wrap(
			mh.ValueMetricHandlerJSON,
			sugar,
			mws...,
		))
		r.Get("/{type}/{name}", wrap(
			mh.GetMetricHandler,
			sugar,
			mws...,
		))

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
