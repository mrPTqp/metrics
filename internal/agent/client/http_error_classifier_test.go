package agent

import (
	"errors"
	"net"
	"net/url"
	"testing"

	"github.com/mrPTqp/metrics/internal/retry"
)

func TestHTTPErrorClassifier_Classify(t *testing.T) {
	classifier := NewHTTPErrorClassifier()

	timeoutErr := &net.DNSError{
		IsTimeout: true,
	}

	opErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: errors.New("connection refused"),
	}

	dnsErr := &net.DNSError{
		Name: "example.com",
	}

	urlTimeoutErr := &url.Error{
		Op:  "Get",
		URL: "http://example.com",
		Err: timeoutErr,
	}

	urlDNSErr := &url.Error{
		Op:  "Get",
		URL: "http://example.com",
		Err: dnsErr,
	}

	urlNonRetriableErr := &url.Error{
		Op:  "Get",
		URL: "http://example.com",
		Err: errors.New("permanent error"),
	}

	tests := []struct {
		name string
		err  error
		want retry.ErrorClassification
	}{
		{
			name: "nil error",
			err:  nil,
			want: retry.NonRetriable,
		},
		{
			name: "net timeout error",
			err:  timeoutErr,
			want: retry.Retriable,
		},
		{
			name: "net OpError",
			err:  opErr,
			want: retry.Retriable,
		},
		{
			name: "url timeout error",
			err:  urlTimeoutErr,
			want: retry.Retriable,
		},
		{
			name: "url DNS error",
			err:  urlDNSErr,
			want: retry.Retriable,
		},
		{
			name: "url non-retriable",
			err:  urlNonRetriableErr,
			want: retry.NonRetriable,
		},
		{
			name: "generic non-retriable",
			err:  errors.New("some error"),
			want: retry.NonRetriable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifier.Classify(tt.err)
			if got != tt.want {
				t.Errorf("Classify(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

