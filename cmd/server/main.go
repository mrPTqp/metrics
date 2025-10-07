package main

import (
	"log"
	"net/http"

	"github.com/mrPTqp/metrics/internal/handler"
	"github.com/mrPTqp/metrics/internal/service"
	"github.com/mrPTqp/metrics/internal/repository"
)

func main() {
	mux := http.NewServeMux()
	mr := repository.NewMemStorage()
	ms := service.NewMetricsService(mr)
	mh := handler.NewMetricHandler(ms)
	mux.Handle("/update/", mh)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	log.Println("Server running on :8080")
	log.Fatal(srv.ListenAndServe())
}
