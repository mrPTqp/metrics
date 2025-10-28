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

	r := chi.NewRouter()
	r.Get("/", middleware.WithLogging(mh.CollectMetricsHandler, sugar))
	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", middleware.WithLogging(mh.SaveMetricHandler, sugar))
	})
	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", middleware.WithLogging(mh.GetMetricHandler, sugar))
	})

	srv := &http.Server{
		Addr:    address.String(),
		Handler: r,
	}
	sugar.Info("Server running on %s", address.String())
	sugar.Fatal(srv.ListenAndServe())
}
