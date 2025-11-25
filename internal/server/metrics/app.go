package metrics

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/mrPTqp/metrics/internal/server/config"
	"github.com/mrPTqp/metrics/internal/server/handler"
)

func StartMetricsServer(mh *handler.MetricHandler, mws []func(h http.HandlerFunc, sugar *zap.SugaredLogger) http.HandlerFunc, cfg *config.Config, sugar *zap.SugaredLogger) *http.Server {
	r := chi.NewRouter()
	r.Get("/", wrap(mh.CollectMetricsHandler, sugar, mws...))
	if cfg.DatabaseDsn != "" {
		r.Get("/ping", wrap(mh.DBHealthCheckHandler, sugar, mws...))
	}
	r.Route("/update", func(r chi.Router) {
		r.Post("/", wrap(mh.SaveMetricHandlerJSON, sugar, mws...))
		r.Post("/{type}/{name}/{value}", wrap(mh.SaveMetricHandler, sugar, mws...))
	})
	r.Route("/updates", func(r chi.Router) {
		r.Post("/", wrap(mh.SaveMetricsHandlerJSON, sugar, mws...))
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", wrap(mh.ValueMetricHandlerJSON, sugar, mws...))
		r.Get("/{type}/{name}", wrap(mh.GetMetricHandler, sugar, mws...))
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
