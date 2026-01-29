package audit

import (
	"context"

	"go.uber.org/zap"
)

type EventBus struct {
	events     chan AuditEvent
	processors []AuditProcessor
	done       chan struct{}
	logger     *zap.Logger
}

func NewEventBus(processors []AuditProcessor, bufferSize int, logger *zap.Logger) *EventBus {
	bus := &EventBus{
		events:     make(chan AuditEvent, bufferSize),
		processors: processors,
		done:       make(chan struct{}),
		logger:     logger,
	}
	go bus.start()
	return bus
}

func (b *EventBus) Publish(e AuditEvent) {
	select {
	case b.events <- e:
	default:
		b.logger.Warn("event was dropped because event bus is full", zap.Any("event", e))
	}
}

func (b *EventBus) start() {
	defer close(b.done)

	for {
		event, ok := <-b.events
		if !ok {
			return
		}
		for _, processor := range b.processors {
			go b.sendToProcessor(processor, event)
		}
	}
}

func (b *EventBus) sendToProcessor(processor AuditProcessor, event AuditEvent) {
	if err := processor.Write(event); err != nil {
		b.logger.Warn("error send event to audit processor", zap.Any("processor", processor), zap.Error(err))
	}
}

func (b *EventBus) ShutDown(ctx context.Context) error {
	close(b.events)

	select {
	case <-b.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
