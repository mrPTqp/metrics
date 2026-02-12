package scheduler

import (
	"context"
	"sync"
	"time"

	agent "github.com/mrPTqp/metrics/internal/agent/client"
	"go.uber.org/zap"
)

// Планировщик сбора и отправки метрик
type MetricsScheduler struct {
	agent  *agent.MetricsAgent
	logger *zap.Logger
}

// Возвращает новый экземпляр MetricsScheduler
func NewScheduler(agent *agent.MetricsAgent, logger *zap.Logger) *MetricsScheduler {
	return &MetricsScheduler{
		agent:  agent,
		logger: logger,
	}
}

// Запускает процесс сбора и отправки метрик
func (s *MetricsScheduler) Start(ctx context.Context, pollInterval int, reportInterval int, rateLimit int, taskChannelSize int) {
	taskCh := make(chan struct{}, taskChannelSize)
	var wg sync.WaitGroup

	pollTicker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(reportInterval) * time.Second)

	defer func() {
		pollTicker.Stop()
		reportTicker.Stop()
		close(taskCh)
		s.logger.Info("Scheduler stopped: tickers stopped and task channel closed")
	}()

	for i := range rateLimit {
		wg.Add(1)
		go s.worker(i, taskCh, &wg)
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Shutdown signal received, waiting for active workers to finish...")
			wg.Wait()
			s.logger.Info("All workers have stopped. Scheduler shutdown complete.")
			return

		case <-pollTicker.C:
			s.agent.PollMetrics()
			s.agent.PollAdditionalGaugeMetrics()

		case <-reportTicker.C:
			select {
			case taskCh <- struct{}{}:
				s.logger.Debug("Scheduled metrics send")
			default:
				s.logger.Warn("Task queue is full, skipping metrics send")
			}
		}
	}
}

func (s *MetricsScheduler) worker(id int, tasks <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	s.logger.Debug("Worker started and waiting for tasks", zap.Int("worker_id", id))

	for range tasks {
		s.logger.Debug("Worker received task, sending metrics", zap.Int("worker_id", id))
		s.agent.SendMetrics()
		s.logger.Debug("Worker finished sending metrics", zap.Int("worker_id", id))
	}

	s.logger.Debug("Worker shutting down (task channel closed)", zap.Int("worker_id", id))
}
