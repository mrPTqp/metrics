package audit

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

const workerCount = 10

// Шина для публикации событий обработки метрик
type EventBus struct {
	events     chan AuditEvent
	processors []AuditProcessor
	done       chan struct{}
	jobs       chan job
	logger     *zap.Logger
}

type job struct {
	processor AuditProcessor
	event     AuditEvent
}

// Возвращает новый экземпляр EventBus
func NewEventBus(processors []AuditProcessor, bufferSize int, logger *zap.Logger) *EventBus {
	bus := &EventBus{
		events:     make(chan AuditEvent, bufferSize),
		processors: processors,
		done:       make(chan struct{}),
		jobs:       make(chan job, workerCount*2),
		logger:     logger,
	}

	for i := 0; i < workerCount; i++ {
		go bus.worker()
	}

	go bus.start()
	return bus
}

func (b *EventBus) worker() {
	for j := range b.jobs {
		if err := j.processor.Write(j.event); err != nil {
			b.logger.Warn("error sending event to audit processor",
				zap.String("processor", b.processorName(j.processor)),
				zap.Error(err),
				zap.Any("event", j.event))
		}
	}
}

// Публикует событие в шину
func (b *EventBus) Publish(e AuditEvent) error {
	select {
	case b.events <- e:
		return nil
	default:
		return errors.New("audit bus full: event rejected")
	}
}

func (b *EventBus) start() {
	defer close(b.done)

	for event := range b.events {
		for _, processor := range b.processors {
			select {
			case b.jobs <- job{processor: processor, event: event}:
			default:
				b.logger.Warn("audit job queue full, event skipped by processor",
					zap.String("processor", b.processorName(processor)),
					zap.Any("event", event))
			}
		}
	}
}

func (b *EventBus) ShutDown(ctx context.Context) error {
	close(b.events) 

	select {
	case <-b.done:
		close(b.jobs)
		return nil
	case <-ctx.Done():
		close(b.jobs)
		return ctx.Err()
	}
}

func (b *EventBus) processorName(p AuditProcessor) string {
	if named, ok := p.(interface{ Name() string }); ok {
		return named.Name()
	}
	return fmt.Sprintf("%T", p)
}