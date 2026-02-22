package audit

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"

	"go.uber.org/zap"
)

// Обработчик событий аудита с сохранением в файл
type FileAuditProcessor struct {
	ch     chan []byte
	file   *os.File
	done   chan struct{}
	logger *zap.Logger
}

// Возвращает новый экземпляр FileAuditProcessor
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

	defer func() {
		if err := writer.Flush(); err != nil {
			p.logger.Error("failed to flush writer on close", zap.Error(err))
		}
	}()
	defer func() {
		if err := p.file.Close(); err != nil {
			p.logger.Error("failed to close audit file", zap.Error(err))
		}
	}()

	for data := range p.ch {
		_, err := writer.Write(data)
		if err != nil {
			p.logger.Error("failed to write audit event", zap.Error(err))
			continue
		}

		err = writer.Flush()
		if err != nil {
			p.logger.Error("failed to flush audit event", zap.Error(err))
			continue
		}
	}
}

// Записывает событие аудита в файл
func (p *FileAuditProcessor) Write(event AuditEvent) error {
	select {
	case p.ch <- event.marshalWithNewline():
		return nil
	default:
		p.logger.Warn("audit event dropped: channel is full")
		return errors.New("audit event dropped: channel is full")
	}
}

func (e *AuditEvent) marshalWithNewline() []byte {
	data, _ := json.Marshal(e)
	return append(data, '\n')
}

// Останавливает обработчик аудита с сохраниением в файл
func (p *FileAuditProcessor) ShutDown(ctx context.Context) error {
	close(p.ch)

	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
