package scheduler

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	agent "github.com/mrPTqp/metrics/internal/agent/client"
	agentcfg "github.com/mrPTqp/metrics/internal/agent/config"
	agentstorage "github.com/mrPTqp/metrics/internal/agent/storage"
	"github.com/mrPTqp/metrics/internal/logger"
)

// TestMetricsScheduler_worker verifies that worker consumes tasks from the channel
// and calls agent.SendMetrics without panicking.
func TestMetricsScheduler_worker(t *testing.T) {
	log := logger.NewLogger()
	defer func() { _ = log.Sync() }()

	cfg := &agentcfg.Config{} // minimal config; agent tests already cover its behaviour in depth
	mem := agentstorage.NewMemStorage()
	httpClient := &http.Client{
		Timeout: time.Second,
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			// Возвращаем пустой успешный ответ, чтобы агент не делал реальных HTTP-запросов.
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
			}, nil
		}),
	}
	a := agent.NewMetricsAgent(httpClient, cfg, mem, log)

	s := NewScheduler(a, nil, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tasks := make(chan struct{}, 1)

	var wg sync.WaitGroup
	wg.Add(1)
	go s.worker(ctx, 1, tasks, &wg)

	// отправляем одну задачу и закрываем канал
	tasks <- struct{}{}
	close(tasks)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not finish in time")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}


