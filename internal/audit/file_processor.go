package audit

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	"go.uber.org/zap"
)

type FileAuditProcessor struct {
	ch     chan []byte
	file   *os.File
	done   chan struct{}
	logger *zap.Logger
}

func NewFileAuditProcessor(filePath string, logger *zap.Logger) (*FileAuditProcessor, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	processor := &FileAuditProcessor{
		ch:     make(chan []byte),
		file:   file,
		done:   make(chan struct{}),
		logger: logger,
	}

	go processor.writer()
	return processor, nil
}

func (p *FileAuditProcessor) writer() {
	defer close(p.done)
	writer := bufio.NewWriter(p.file)
	defer writer.Flush()
	defer p.file.Close()

	for data := range p.ch {
		_, _ = writer.Write(data)
		_ = writer.Flush()
	}
}

func (p *FileAuditProcessor) Write(event AuditEvent) error {
	select {
	case p.ch <- event.marshalWithNewline():
		return nil
	default:
		p.logger.Warn("event was dropped because file channel is full")
	}
	return nil
}

func (e *AuditEvent) marshalWithNewline() []byte {
	data, _ := json.Marshal(e)
	return append(data, '\n')
}

func (p *FileAuditProcessor) ShutDown(ctx context.Context) error {
	close(p.ch)

	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
