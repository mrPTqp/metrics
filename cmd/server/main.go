// cmd/server/main.go
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

	// Создаём общий контекст приложения
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Передаём контекст в бутстрап — например, для восстановления из файла
	components := bootstrapper.MustRun(ctx)
	if components == nil {
		sugar.Fatal("failed to bootstrap application")
	}

	application := app.NewApp(components)

	// Запускаем приложение с общим контекстом
	go application.RunWithContext(ctx)

	sugar.Infoln("Application started")

	<-ctx.Done()
	sugar.Infoln("Shutdown signal received")

	// Контекст для graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	application.Shutdown(shutdownCtx)
	sugar.Infoln("Application stopped")
}
