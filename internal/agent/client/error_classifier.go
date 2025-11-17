package agent

import (
	"errors"
	"net"
	"net/url"
)

type HTTPErrorClassification int

const (
	NonRetriable HTTPErrorClassification = iota
	Retriable
)

type HTTPErrorClassifier struct{}

func NewHTTPErrorClassifier() *HTTPErrorClassifier {
	return &HTTPErrorClassifier{}
}

func (c *HTTPErrorClassifier) Classify(err error) HTTPErrorClassification {
	if err == nil {
		return NonRetriable
	}

	// Проверка на сетевые ошибки (timeout, connection refused и т.п.)
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return Retriable
	}

	// Проверка на операционные ошибки (например, connection refused)
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return Retriable
	}

	// URL-ошибки, вызванные сетевыми проблемами
	if urlErr, ok := err.(*url.Error); ok {
		var urlOpErr *net.OpError
		var dnsErr *net.DNSError
		if errors.As(urlErr.Err, &urlOpErr) || urlErr.Timeout() || errors.As(urlErr.Err, &dnsErr) {
			return Retriable
		}
		return NonRetriable
	}

	return NonRetriable
}
