package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestDbHealthCheckHandler_PingSuccess(t *testing.T) {
	logger := zap.NewNop().Sugar()
	mock := &mockMetricsService{
		pingFunc: func() bool {
			return true
		},
	}

	mh := NewMetricHandler(mock, logger)
	r := chi.NewRouter()
	r.Get("/ping", mh.DBHealthCheckHandler)

	req := httptest.NewRequest("GET", "/ping", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestDbHealthCheckHandler_PingFailure(t *testing.T) {
	logger := zap.NewNop().Sugar()
	mock := &mockMetricsService{
		pingFunc: func() bool {
			return false
		},
	}

	mh := NewMetricHandler(mock, logger)
	r := chi.NewRouter()
	r.Get("/ping", mh.DBHealthCheckHandler)

	req := httptest.NewRequest("GET", "/ping", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDbHealthCheckHandler_DifferentHTTPMethods(t *testing.T) {
	logger := zap.NewNop().Sugar()
	mock := &mockMetricsService{
		pingFunc: func() bool {
			return true
		},
	}

	mh := NewMetricHandler(mock, logger)
	r := chi.NewRouter()
	r.Get("/ping", mh.DBHealthCheckHandler)

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/ping", nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if method == "GET" {
				assert.Equal(t, http.StatusOK, rec.Code)
			} else {
				assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
			}
		})
	}
}

func TestDbHealthCheckHandler_MultipleRequests(t *testing.T) {
	logger := zap.NewNop().Sugar()
	pingCount := 0

	mock := &mockMetricsService{
		pingFunc: func() bool {
			pingCount++
			return pingCount%2 == 1
		},
	}

	mh := NewMetricHandler(mock, logger)
	r := chi.NewRouter()
	r.Get("/ping", mh.DBHealthCheckHandler)

	req1 := httptest.NewRequest("GET", "/ping", nil)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	req2 := httptest.NewRequest("GET", "/ping", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusInternalServerError, rec2.Code)

	assert.Equal(t, 2, pingCount)
}

func TestDbHealthCheckHandler_ConcurrentRequests(t *testing.T) {
	logger := zap.NewNop().Sugar()
	pingCount := 0

	mock := &mockMetricsService{
		pingFunc: func() bool {
			pingCount++
			return true
		},
	}

	mh := NewMetricHandler(mock, logger)
	r := chi.NewRouter()
	r.Get("/ping", mh.DBHealthCheckHandler)

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/ping", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	assert.Equal(t, 10, pingCount)
}

func TestDbHealthCheckHandler_NoBody(t *testing.T) {
	logger := zap.NewNop().Sugar()
	mock := &mockMetricsService{
		pingFunc: func() bool {
			return true
		},
	}

	mh := NewMetricHandler(mock, logger)
	r := chi.NewRouter()
	r.Get("/ping", mh.DBHealthCheckHandler)

	req := httptest.NewRequest("GET", "/ping", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	bodyLen := rec.Body.Len()
	assert.Equal(t, 0, bodyLen, "Response body should be empty")
}

func TestDbHealthCheckHandler_ServiceNil(t *testing.T) {
	logger := zap.NewNop().Sugar()
	mock := &mockMetricsService{
		pingFunc: func() bool {
			return false
		},
	}

	mh := NewMetricHandler(mock, logger)
	r := chi.NewRouter()
	r.Get("/ping", mh.DBHealthCheckHandler)

	req := httptest.NewRequest("GET", "/ping", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.NotNil(t, res)
}

func TestNewSupportHandler_CreatesHandler(t *testing.T) {
	logger := zap.NewNop().Sugar()
	mock := &mockMetricsService{
		pingFunc: func() bool {
			return true
		},
	}

	handler := NewSupportHandler(mock, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, mock, handler.service)
	assert.Equal(t, logger, handler.logger)
}

func TestNewSupportHandler_WithNilService(t *testing.T) {
	logger := zap.NewNop().Sugar()

	handler := NewSupportHandler(nil, logger)

	assert.NotNil(t, handler)
	assert.Nil(t, handler.service)
	assert.Equal(t, logger, handler.logger)
}

func TestNewSupportHandler_WithNilLogger(t *testing.T) {
	mock := &mockMetricsService{
		pingFunc: func() bool {
			return true
		},
	}

	handler := NewSupportHandler(mock, nil)

	assert.NotNil(t, handler)
	assert.Equal(t, mock, handler.service)
	assert.Nil(t, handler.logger)
}

func TestNewSupportHandler_CanUseDbHealthCheckHandler(t *testing.T) {
	logger := zap.NewNop().Sugar()
	mock := &mockMetricsService{
		pingFunc: func() bool {
			return true
		},
	}

	handler := NewSupportHandler(mock, logger)
	r := chi.NewRouter()
	r.Get("/ping", handler.DBHealthCheckHandler)

	req := httptest.NewRequest("GET", "/ping", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDbHealthCheckHandler_PingReturnsTrueMultipleTimes(t *testing.T) {
	logger := zap.NewNop().Sugar()
	callCount := 0

	mock := &mockMetricsService{
		pingFunc: func() bool {
			callCount++
			return true
		},
	}

	mh := NewMetricHandler(mock, logger)
	r := chi.NewRouter()
	r.Get("/ping", mh.DBHealthCheckHandler)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/ping", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	assert.Equal(t, 5, callCount)
}

func TestDbHealthCheckHandler_PingReturnsFalseMultipleTimes(t *testing.T) {
	logger := zap.NewNop().Sugar()
	callCount := 0

	mock := &mockMetricsService{
		pingFunc: func() bool {
			callCount++
			return false
		},
	}

	mh := NewMetricHandler(mock, logger)
	r := chi.NewRouter()
	r.Get("/ping", mh.DBHealthCheckHandler)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/ping", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}

	assert.Equal(t, 5, callCount)
}

func TestDbHealthCheckHandler_Headers(t *testing.T) {
	logger := zap.NewNop().Sugar()
	mock := &mockMetricsService{
		pingFunc: func() bool {
			return true
		},
	}

	mh := NewMetricHandler(mock, logger)
	r := chi.NewRouter()
	r.Get("/ping", mh.DBHealthCheckHandler)

	req := httptest.NewRequest("GET", "/ping", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.NotNil(t, res.Header)
}

func TestDbHealthCheckHandler_ServiceMethodCalled(t *testing.T) {
	logger := zap.NewNop().Sugar()
	pingCalled := false

	mock := &mockMetricsService{
		pingFunc: func() bool {
			pingCalled = true
			return true
		},
	}

	mh := NewMetricHandler(mock, logger)
	r := chi.NewRouter()
	r.Get("/ping", mh.DBHealthCheckHandler)

	req := httptest.NewRequest("GET", "/ping", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.True(t, pingCalled)
}
