package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/handler"
	"github.com/mrPTqp/metrics/internal/middleware"
	"github.com/mrPTqp/metrics/internal/repository"
	"github.com/mrPTqp/metrics/internal/service"
)

type NetAddress struct {
	Host string
	Port int
}

var address NetAddress = NetAddress{"localhost", 8080}

func main() {
	parseFlags()
	parseEnvs()

	var sugar *zap.SugaredLogger
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugar = logger.Sugar()

	mr := repository.NewMemStorage(sugar)
	ms := service.NewMetricsService(mr, sugar)
	mh := handler.NewMetricHandler(ms, sugar)

	mws := []func(h http.HandlerFunc, sugar *zap.SugaredLogger) http.HandlerFunc{
		middleware.LoggingMiddleware,
		middleware.GzipMiddleware,
	}
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
		Addr:    address.String(),
		Handler: r,
	}
	sugar.Info("Server running on %s", address.String())
	sugar.Fatal(srv.ListenAndServe())
}

func wrap(h http.HandlerFunc, logger *zap.SugaredLogger, mwFuncs ...func(http.HandlerFunc, *zap.SugaredLogger) http.HandlerFunc) http.HandlerFunc {
	for i := len(mwFuncs) - 1; i >= 0; i-- {
		h = mwFuncs[i](h, logger)
	}
	return h
}
