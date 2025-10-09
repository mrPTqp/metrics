// internal/handler/handler_test.go
package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrPTqp/metrics/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMetricsService struct {
	gaugeCalls    []string
	gaugeValues   []float64
	counterCalls  []string
	counterValues []int64
	err           error
}

func (m *mockMetricsService) ProcessGaugeMetric(name string, value float64) error {
	m.gaugeCalls = append(m.gaugeCalls, name)
	m.gaugeValues = append(m.gaugeValues, value)
	return m.err
}

func (m *mockMetricsService) ProcessCounterMetric(name string, value int64) error {
	m.counterCalls = append(m.counterCalls, name)
	m.counterValues = append(m.counterValues, value)
	return m.err
}

// Проверка, что mockMetricsService реализует MetricService
var _ service.MetricsService = (*mockMetricsService)(nil)

type want struct {
	statusCode   int
	contentType  string
	errorMessage string
	gaugeCall    bool
	gaugeName    string
	gaugeValue   float64
	counterCall  bool
	counterName  string
	counterValue int64
}

func TestMetricHandler(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		ms     *mockMetricsService
		want   want
	}{
		{
			name:   "invalid method",
			method: http.MethodGet,
			path:   "/update/gauge/memory/123.45",
			ms:     &mockMetricsService{},
			want: want{
				statusCode:   http.StatusMethodNotAllowed,
				contentType:  "text/plain; charset=utf-8",
				errorMessage: "Method not allowed",
			},
		},
		{
			name:   "invalid URL format (too few parts)",
			method: http.MethodPost,
			path:   "/update/gauge/memory",
			ms:     &mockMetricsService{},
			want: want{
				statusCode:   http.StatusNotFound,
				contentType:  "text/plain; charset=utf-8",
				errorMessage: "Invalid URL format. Expected: /update/<type>/<name>/<value>",
			},
		},
		{
			name:   "invalid URL format (too many parts)",
			method: http.MethodPost,
			path:   "/update/gauge/memory/123.45/extra",
			ms:     &mockMetricsService{},
			want: want{
				statusCode:   http.StatusNotFound,
				contentType:  "text/plain; charset=utf-8",
				errorMessage: "Invalid URL format. Expected: /update/<type>/<name>/<value>",
			},
		},
		{
			name:   "invalid metric type",
			method: http.MethodPost,
			path:   "/update/int/memory/123.45",
			ms:     &mockMetricsService{},
			want: want{
				statusCode:   http.StatusBadRequest,
				contentType:  "text/plain; charset=utf-8",
				errorMessage: "Invalid metric type",
			},
		},
		{
			name:   "invalid metric name (empty)",
			method: http.MethodPost,
			path:   "/update/gauge//123.45",
			ms:     &mockMetricsService{},
			want: want{
				statusCode:   http.StatusBadRequest,
				contentType:  "text/plain; charset=utf-8",
				errorMessage: "Invalid metric name",
			},
		},
		{
			name:   "invalid gauge value (non-float)",
			method: http.MethodPost,
			path:   "/update/gauge/memory/abc",
			ms:     &mockMetricsService{},
			want: want{
				statusCode:   http.StatusBadRequest,
				contentType:  "text/plain; charset=utf-8",
				errorMessage: "Invalid gauge value",
			},
		},
		{
			name:   "invalid counter value (non-int)",
			method: http.MethodPost,
			path:   "/update/counter/requests/abc",
			ms:     &mockMetricsService{},
			want: want{
				statusCode:   http.StatusBadRequest,
				contentType:  "text/plain; charset=utf-8",
				errorMessage: "Invalid counter value",
			},
		},
		{
			name:   "service error during gauge processing",
			method: http.MethodPost,
			path:   "/update/gauge/memory/123.45",
			ms: &mockMetricsService{
				err: assert.AnError,
			},
			want: want{
				statusCode:   http.StatusBadRequest,
				contentType:  "text/plain; charset=utf-8",
				errorMessage: assert.AnError.Error(),
			},
		},
		{
			name:   "valid gauge update",
			method: http.MethodPost,
			path:   "/update/gauge/memory/123.45",
			ms:     &mockMetricsService{},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				gaugeCall:   true,
				gaugeName:   "memory",
				gaugeValue:  123.45,
			},
		},
		{
			name:   "valid counter update",
			method: http.MethodPost,
			path:   "/update/counter/requests/42",
			ms:     &mockMetricsService{},
			want: want{
				statusCode:   http.StatusOK,
				contentType:  "text/plain; charset=utf-8",
				counterCall:  true,
				counterName:  "requests",
				counterValue: 42,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, nil)
			rr := httptest.NewRecorder()
			handler := NewMetricHandler(test.ms)
			handler.ServeHTTP(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.statusCode, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

			if res.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(res.Body)
				assert.Contains(t, string(body), test.want.errorMessage)
			} else {
				if test.want.gaugeCall {
					require.Len(t, test.ms.gaugeCalls, 1)
					assert.Equal(t, test.want.gaugeName, test.ms.gaugeCalls[0])
					assert.InDelta(t, test.want.gaugeValue, test.ms.gaugeValues[0], 0.0001)
				}
				if test.want.counterCall {
					require.Len(t, test.ms.counterCalls, 1)
					assert.Equal(t, test.want.counterName, test.ms.counterCalls[0])
					assert.Equal(t, test.want.counterValue, test.ms.counterValues[0])
				}
			}
		})
	}
}
