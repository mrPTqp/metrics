package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/backup"
	"github.com/mrPTqp/metrics/internal/handler"
	"github.com/mrPTqp/metrics/internal/middleware"
	"github.com/mrPTqp/metrics/internal/repository"
	"github.com/mrPTqp/metrics/internal/scheduler"
	"github.com/mrPTqp/metrics/internal/service"
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

	mr := repository.NewMemStorage(sugar)
	ms := service.NewMetricsService(mr, sugar, syncBackup)
	mh := handler.NewMetricHandler(ms, sugar)

	mws := []func(h http.HandlerFunc, sugar *zap.SugaredLogger) http.HandlerFunc{
		middleware.LoggingMiddleware,
		middleware.GzipMiddleware,
	}
	startMetricsServer(mh, mws, cfg, sugar)

	producer, err := backup.NewProducer(cfg.File)
	if err != nil {
		sugar.Fatal(err)
	}
	defer producer.Close()

	b := backup.NewFileBackuper(ms, producer, sugar)
	sc := scheduler.NewScheduler(b, sugar)
	sc.Start(cfg.StoreInterval, cfg.File)
	defer os.Remove(cfg.File)
}

func startMetricsServer(mh *handler.MetricHandler, mws []func(h http.HandlerFunc, sugar *zap.SugaredLogger) http.HandlerFunc, config *Config, sugar *zap.SugaredLogger) {
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
		Addr:    config.Address.String(),
		Handler: r,
	}
	sugar.Infof("Server running on %s", config.Address.String())
	sugar.Fatal(srv.ListenAndServe())
}

func wrap(h http.HandlerFunc, logger *zap.SugaredLogger, mwFuncs ...func(http.HandlerFunc, *zap.SugaredLogger) http.HandlerFunc) http.HandlerFunc {
	for i := len(mwFuncs) - 1; i >= 0; i-- {
		h = mwFuncs[i](h, logger)
	}
	return h
}
