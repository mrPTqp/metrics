package agent

import (
	"errors"
	"net"
	"net/url"

	"github.com/mrPTqp/metrics/internal/retry"
)

// Классификтор HTTP ошибок
type HTTPErrorClassifier struct{}

// Возвращает новый экземпляр HTTPErrorClassifier
func NewHTTPErrorClassifier() *HTTPErrorClassifier {
	return &HTTPErrorClassifier{}
}

// Принимает решение относится ли ошибка к тем, по которым стоит повторить запрос
func (c *HTTPErrorClassifier) Classify(err error) retry.ErrorClassification {
	if err == nil {
		return retry.NonRetriable
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return retry.Retriable
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return retry.Retriable
	}

	if urlErr, ok := err.(*url.Error); ok {
		var urlOpErr *net.OpError
		var dnsErr *net.DNSError
		if errors.As(urlErr.Err, &urlOpErr) || urlErr.Timeout() || errors.As(urlErr.Err, &dnsErr) {
			return retry.Retriable
		}
		return retry.NonRetriable
	}

	return retry.NonRetriable
}
