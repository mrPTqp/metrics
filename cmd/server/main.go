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
	mr := repository.NewMemStorage()
	ms := service.NewMetricsService(mr)
	mh := handler.NewMetricHandler(ms)

	r := chi.NewRouter()
	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", mh.SaveMetricHandler)
		})
	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", mh.GetMetricHandler)
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	log.Println("Server running on :8080")
	log.Fatal(srv.ListenAndServe())
}
