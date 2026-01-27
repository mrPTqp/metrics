package retry

import (
	"context"
	"time"
)

// ErrorClassifier интерфейс для классификации ошибок: повторяемая или нет.
type ErrorClassifier interface {
	Classify(error) ErrorClassification
}

// ErrorClassification определяет, следует ли повторять операцию.
type ErrorClassification int

const (
	NonRetriable ErrorClassification = iota
	Retriable
)

func DoWithRetry(
	ctx context.Context,
	classifier ErrorClassifier,
	operation func() error,
	maxRetries int,
	initialDelay time.Duration,
) error {
	var err error
	delay := initialDelay

	for i := 0; i <= maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		if classifier != nil {
			if classification := classifier.Classify(err); classification == NonRetriable {
				return err
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if i == maxRetries {
			break
		}

		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}

		delay += 2 * time.Second
	}

	return err
}
