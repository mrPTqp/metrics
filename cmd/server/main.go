package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrPTqp/metrics/internal/logger"
	"github.com/mrPTqp/metrics/internal/server/config"
	"github.com/mrPTqp/metrics/internal/server/app"
	"github.com/mrPTqp/metrics/internal/server/bootstrap"
)

func main() {
	sugar := logger.NewSugarLogger()

	cfg := config.LoadConfig()
	sugar.Infow("configuration created", "config", cfg)

	bootstrapper := bootstrap.NewBootstrapper(cfg, sugar)
	components := bootstrapper.MustRun()

	application := metrics.NewApp(components)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go application.Run()

	<-ctx.Done()
	sugar.Infoln("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	application.Shutdown(shutdownCtx)
	sugar.Infoln("Application stopped")
}
