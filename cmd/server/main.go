package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mrPTqp/metrics/internal/handler"
	"github.com/mrPTqp/metrics/internal/repository"
	"github.com/mrPTqp/metrics/internal/service"
)

func main() {
	parseFlags()

	mr := repository.NewMemStorage()
	ms := service.NewMetricsService(mr)
	mh := handler.NewMetricHandler(ms)

	r := chi.NewRouter()
	r.Get("/", mh.CollectMetricsHandler)
	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", mh.SaveMetricHandler)
	})
	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", mh.GetMetricHandler)
	})

	srv := &http.Server{
		Addr:    address.String(),
		Handler: r,
	}
	log.Printf("Server running on %s", address.String())
	log.Fatal(srv.ListenAndServe())
}
