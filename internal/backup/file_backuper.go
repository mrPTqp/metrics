package backup

import (
	"encoding/json"
	"os"

	"github.com/mrPTqp/metrics/internal/models"
	"github.com/mrPTqp/metrics/internal/service"
	"go.uber.org/zap"
)

type FileBackuper struct {
	Service  *service.BaseMetricService
	Producer *Producer
	Logger   *zap.SugaredLogger
}

func NewFileBackuper(service *service.BaseMetricService, producer *Producer, logger *zap.SugaredLogger) *FileBackuper {
	return &FileBackuper{
		Service:  service,
		Producer: producer,
		Logger:   logger,
	}
}

func (b *FileBackuper) Backup() {
	gauges, counters := b.Service.ListAllMetrics()

	for name, value := range gauges {
		metric := models.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &value,
		}
		if err := b.Producer.WriteEvent(&metric); err != nil {
			b.Logger.Fatal(err)
		}
	}

	for name, value := range counters {
		metric := models.Metrics{
			ID:    name,
			MType: "counter",
			Value: nil,
			Delta: &value,
		}
		if err := b.Producer.WriteEvent(&metric); err != nil {
			b.Logger.Fatal(err)
		}
	}
}

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(fileName string) (*Producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Producer) WriteEvent(event *models.Metrics) error {
	return p.encoder.Encode(&event)
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(fileName string) (*Consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *Consumer) ReadEvent() (*models.Metrics, error) {
	event := &models.Metrics{}
	if err := c.decoder.Decode(&event); err != nil {
		return nil, err
	}

	return event, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}
