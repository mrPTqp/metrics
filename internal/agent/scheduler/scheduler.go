package scheduler

import (
	"time"

	"github.com/mrPTqp/metrics/internal/agent/client"
	"go.uber.org/zap"
)

type MetricsScheduler struct {
	agent  *agent.MetricsAgent
	logger *zap.SugaredLogger
}

func NewScheduler(agent *agent.MetricsAgent, logger *zap.SugaredLogger) *MetricsScheduler {
	return &MetricsScheduler{
		agent:  agent,
		logger: logger,
	}
}

func (s *MetricsScheduler) Start(poolInterval int, reportInterval int, rateLimit int, taskChannelSize int) {
	taskCh := make(chan struct{}, taskChannelSize)

	poolTicker := time.NewTicker(time.Duration(poolInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(reportInterval) * time.Second)

	defer func() {
		poolTicker.Stop()
		reportTicker.Stop()
		close(taskCh)
	}()

	for i := range rateLimit {
		go s.worker(i, taskCh)
	}

	for {
		select {
		case <-poolTicker.C:
			go s.agent.PoolMetrics()
			go s.agent.PoolAdditionalGaugeMetrics()

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

func (s *MetricsScheduler) worker(id int, tasks <-chan struct{}) {
	s.logger.Debugf("Worker %d: started and waiting for tasks", id)

	for range tasks {
		s.logger.Debugf("Worker %d: received task, sending metrics", id)
		s.agent.SendMetrics()
		s.logger.Debugf("Worker %d: metrics sent", id)
	}

	s.logger.Debugf("Worker %d: shutting down (task channel closed)", id)
}
