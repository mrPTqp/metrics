package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type HTTPAuditProcessor struct {
	ch     chan AuditEvent
	client *http.Client
	done   chan struct{}
	URL    string
	logger *zap.Logger
}

func NewHTTPAuditProcessor(URL string, logger *zap.Logger) *HTTPAuditProcessor {
	processor := &HTTPAuditProcessor{
		ch:   make(chan AuditEvent),
		done: make(chan struct{}),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		URL: URL,
		logger: logger,
	}

	go processor.sender()
	return processor
}

func (p *HTTPAuditProcessor) Write(event AuditEvent) error {
	select {
	case p.ch <- event:
		p.send(event)
	default:
		p.logger.Warn("event was dropped because http channel is full")
	}
	return nil
}

func (p *HTTPAuditProcessor) sender() {
	defer close(p.done)
	for event := range p.ch {
		p.send(event)
	}
}

func (p *HTTPAuditProcessor) send(event AuditEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		p.logger.Warn("error marshal audit event for http request", zap.Error(err))
		return
	}
	req, err := http.NewRequest("POST", p.URL, bytes.NewReader(data))
	if err != nil {
		p.logger.Warn("error build request for audit event for http request", zap.Error(err))
		return
	}
	resp, err := p.client.Do(req)
	if err != nil {
		p.logger.Warn("error response for http audit event", zap.Error(err))
		return
	}
	resp.Body.Close()
}

func (p *HTTPAuditProcessor) ShutDown(ctx context.Context) error {
	close(p.ch)

	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
