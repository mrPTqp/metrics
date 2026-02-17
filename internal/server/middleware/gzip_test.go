package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestCompressWriter_Header(t *testing.T) {
	logger := zap.NewNop()
	w := httptest.NewRecorder()
	cw := newCompressWriter(w, true, logger)

	header := cw.Header()
	if header == nil {
		t.Fatal("Header() returned nil")
	}

	header.Set("Test-Header", "test-value")
	if w.Header().Get("Test-Header") != "test-value" {
		t.Error("Header() should return underlying ResponseWriter's header")
	}
}

func TestCompressWriter_Write(t *testing.T) {
	tests := []struct {
		name           string
		supportsGzip   bool
		contentType    string
		statusCode     int
		data           []byte
		shouldCompress bool
	}{
		{
			name:           "compress JSON with gzip support",
			supportsGzip:   true,
			contentType:    "application/json",
			statusCode:     http.StatusOK,
			data:           []byte(`{"test": "data"}`),
			shouldCompress: true,
		},
		{
			name:           "no compression without gzip support",
			supportsGzip:   false,
			contentType:    "application/json",
			statusCode:     http.StatusOK,
			data:           []byte(`{"test": "data"}`),
			shouldCompress: false,
		},
		{
			name:           "compress text/html",
			supportsGzip:   true,
			contentType:    "text/html",
			statusCode:     http.StatusOK,
			data:           []byte("<html><body>test</body></html>"),
			shouldCompress: true,
		},
		{
			name:           "compress text/plain",
			supportsGzip:   true,
			contentType:    "text/plain",
			statusCode:     http.StatusOK,
			data:           []byte("plain text"),
			shouldCompress: true,
		},
		{
			name:           "compress text/css",
			supportsGzip:   true,
			contentType:    "text/css",
			statusCode:     http.StatusOK,
			data:           []byte("body { color: red; }"),
			shouldCompress: true,
		},
		{
			name:           "compress application/javascript",
			supportsGzip:   true,
			contentType:    "application/javascript",
			statusCode:     http.StatusOK,
			data:           []byte("console.log('test');"),
			shouldCompress: true,
		},
		{
			name:           "no compression for error status",
			supportsGzip:   true,
			contentType:    "application/json",
			statusCode:     http.StatusNotFound,
			data:           []byte(`{"error": "not found"}`),
			shouldCompress: false,
		},
		{
			name:           "no compression for unsupported content type",
			supportsGzip:   true,
			contentType:    "image/png",
			statusCode:     http.StatusOK,
			data:           []byte("binary data"),
			shouldCompress: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			logger := zap.NewNop()
			cw := newCompressWriter(w, tt.supportsGzip, logger)

			w.Header().Set("Content-Type", tt.contentType)
			cw.WriteHeader(tt.statusCode)

			n, err := cw.Write(tt.data)
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}
			if n != len(tt.data) {
				t.Errorf("Write() wrote %d bytes, want %d", n, len(tt.data))
			}

			if err := cw.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}

			result := w.Body.Bytes()
			contentEncoding := w.Header().Get("Content-Encoding")

			if tt.shouldCompress {
				if contentEncoding != "gzip" {
					t.Errorf("expected Content-Encoding: gzip, got: %s", contentEncoding)
				}

				gr, err := gzip.NewReader(bytes.NewReader(result))
				if err != nil {
					t.Fatalf("failed to create gzip reader: %v", err)
				}
				defer func() { _ = gr.Close() }()

				decompressed, err := io.ReadAll(gr)
				if err != nil {
					t.Fatalf("failed to decompress: %v", err)
				}

				if !bytes.Equal(decompressed, tt.data) {
					t.Errorf("decompressed data doesn't match original")
				}

				if w.Header().Get("Content-Length") != "" {
					t.Error("Content-Length should be deleted for compressed responses")
				}
			} else {
				if contentEncoding == "gzip" {
					t.Error("should not set Content-Encoding: gzip")
				}
				if !bytes.Equal(result, tt.data) {
					t.Errorf("data should not be compressed, got: %v, want: %v", result, tt.data)
				}
			}
		})
	}
}

func TestCompressWriter_WriteHeader(t *testing.T) {
	tests := []struct {
		name           string
		supportsGzip   bool
		contentType    string
		statusCode     int
		shouldCompress bool
	}{
		{
			name:           "should compress JSON",
			supportsGzip:   true,
			contentType:    "application/json",
			statusCode:     http.StatusOK,
			shouldCompress: true,
		},
		{
			name:           "should not compress without gzip support",
			supportsGzip:   false,
			contentType:    "application/json",
			statusCode:     http.StatusOK,
			shouldCompress: false,
		},
		{
			name:           "should not compress error status",
			supportsGzip:   true,
			contentType:    "application/json",
			statusCode:     http.StatusInternalServerError,
			shouldCompress: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			logger := zap.NewNop()
			cw := newCompressWriter(w, tt.supportsGzip, logger)

			w.Header().Set("Content-Type", tt.contentType)
			cw.WriteHeader(tt.statusCode)

			if cw.headerWritten != true {
				t.Error("headerWritten should be true after WriteHeader")
			}

			contentEncoding := w.Header().Get("Content-Encoding")
			if tt.shouldCompress {
				if contentEncoding != "gzip" {
					t.Errorf("expected Content-Encoding: gzip, got: %s", contentEncoding)
				}
			} else {
				if contentEncoding == "gzip" {
					t.Error("should not set Content-Encoding: gzip")
				}
			}

			if w.Code != tt.statusCode {
				t.Errorf("expected status code %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestCompressWriter_WriteHeader_Idempotent(t *testing.T) {
	w := httptest.NewRecorder()
	logger := zap.NewNop()
	cw := newCompressWriter(w, true, logger)

	w.Header().Set("Content-Type", "application/json")
	cw.WriteHeader(http.StatusOK)
	cw.WriteHeader(http.StatusNotFound)

	if w.Code != http.StatusOK {
		t.Errorf("WriteHeader should be idempotent, expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCompressWriter_Write_CallsWriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	logger := zap.NewNop()
	cw := newCompressWriter(w, true, logger)

	w.Header().Set("Content-Type", "application/json")
	_, _ = cw.Write([]byte("test"))

	if !cw.headerWritten {
		t.Error("Write() should call WriteHeader() automatically")
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCompressWriter_Close(t *testing.T) {
	tests := []struct {
		name           string
		shouldCompress bool
		headerWritten  bool
		expectError    bool
	}{
		{
			name:           "close compressed writer",
			shouldCompress: true,
			headerWritten:  true,
			expectError:    false,
		},
		{
			name:           "close non-compressed writer",
			shouldCompress: false,
			headerWritten:  true,
			expectError:    false,
		},
		{
			name:           "close before header written",
			shouldCompress: true,
			headerWritten:  false,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			logger := zap.NewNop()
			cw := newCompressWriter(w, true, logger)
			cw.shouldCompress = tt.shouldCompress
			cw.headerWritten = tt.headerWritten

			if tt.headerWritten {
				cw.WriteHeader(http.StatusOK)
			}

			err := cw.Close()
			if (err != nil) != tt.expectError {
				t.Errorf("Close() error = %v, expectError = %v", err, tt.expectError)
			}
		})
	}
}

func TestCompressReader_Read(t *testing.T) {
	originalData := []byte("test data for compression")
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write(originalData)
	if err != nil {
		t.Fatalf("failed to write compressed data: %v", err)
	}
	if err2 := gw.Close(); err2 != nil {
		t.Fatalf("failed to close gzip writer: %v", err2)
	}
	compressedData := buf.Bytes()

	r := io.NopCloser(bytes.NewReader(compressedData))
	cr, err := newCompressReader(r)
	if err != nil {
		t.Fatalf("newCompressReader() error = %v", err)
	}

	decompressed, err := io.ReadAll(cr)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if !bytes.Equal(decompressed, originalData) {
		t.Errorf("decompressed data = %v, want %v", decompressed, originalData)
	}
}

func TestCompressReader_Close(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, _ = gw.Write([]byte("test"))
	_ = gw.Close()

	r := io.NopCloser(bytes.NewReader(buf.Bytes()))
	cr, err := newCompressReader(r)
	if err != nil {
		t.Fatalf("newCompressReader() error = %v", err)
	}

	err = cr.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	err = cr.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

func TestCompressReader_InvalidGzipData(t *testing.T) {
	invalidData := []byte("not a valid gzip stream")
	r := io.NopCloser(bytes.NewReader(invalidData))

	cr, err := newCompressReader(r)
	if err == nil {
		_ = cr.Close()
		t.Error("expected error for invalid gzip data, got nil")
	}
}

func TestGzipMiddleware_CompressResponse(t *testing.T) {
	testData := []byte(`{"message": "test"}`)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(testData)
	})

	middleware := GzipMiddleware(handler)

	tests := []struct {
		name           string
		acceptEncoding string
		shouldCompress bool
	}{
		{
			name:           "compresses when Accept-Encoding contains gzip",
			acceptEncoding: "gzip, deflate",
			shouldCompress: true,
		},
		{
			name:           "compresses when Accept-Encoding is gzip",
			acceptEncoding: "gzip",
			shouldCompress: true,
		},
		{
			name:           "does not compress without Accept-Encoding",
			acceptEncoding: "",
			shouldCompress: false,
		},
		{
			name:           "does not compress with other encoding",
			acceptEncoding: "deflate",
			shouldCompress: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}

			w := httptest.NewRecorder()
			middleware.ServeHTTP(w, req)

			contentEncoding := w.Header().Get("Content-Encoding")
			if tt.shouldCompress {
				if contentEncoding != "gzip" {
					t.Errorf("expected Content-Encoding: gzip, got: %s", contentEncoding)
				}

				gr, err := gzip.NewReader(w.Body)
				if err != nil {
					t.Fatalf("failed to create gzip reader: %v", err)
				}
				defer func() { _ = gr.Close() }()

				decompressed, err := io.ReadAll(gr)
				if err != nil {
					t.Fatalf("failed to decompress: %v", err)
				}

				if !bytes.Equal(decompressed, testData) {
					t.Errorf("decompressed data doesn't match original")
				}
			} else {
				if contentEncoding == "gzip" {
					t.Error("should not set Content-Encoding: gzip")
				}
				if !bytes.Equal(w.Body.Bytes(), testData) {
					t.Errorf("data should not be compressed")
				}
			}
		})
	}
}

func TestGzipMiddleware_DecompressRequest(t *testing.T) {
	originalData := []byte(`{"message": "test"}`)

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, _ = gw.Write(originalData)
	_ = gw.Close()
	compressedData := buf.Bytes()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		if !bytes.Equal(body, originalData) {
			t.Errorf("decompressed body = %v, want %v", body, originalData)
		}

		w.WriteHeader(http.StatusOK)
	})

	middleware := GzipMiddleware(handler)

	tests := []struct {
		name             string
		contentEncoding  string
		body             []byte
		shouldDecompress bool
	}{
		{
			name:             "decompresses when Content-Encoding is gzip",
			contentEncoding:  "gzip",
			body:             compressedData,
			shouldDecompress: true,
		},
		{
			name:             "does not decompress without Content-Encoding",
			contentEncoding:  "",
			body:             originalData,
			shouldDecompress: false,
		},
		{
			name:             "does not decompress with other encoding",
			contentEncoding:  "deflate",
			body:             originalData,
			shouldDecompress: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", bytes.NewReader(tt.body))
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}

			w := httptest.NewRecorder()
			middleware.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
			}
		})
	}
}

func TestGzipMiddleware_InvalidGzipBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with invalid gzip data")
	})

	middleware := GzipMiddleware(handler)

	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("invalid gzip data")))
	req.Header.Set("Content-Encoding", "gzip")

	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestGzipMiddleware_CompressAndDecompress(t *testing.T) {
	originalRequestData := []byte(`{"request": "data"}`)
	originalResponseData := []byte(`{"response": "data"}`)

	var reqBuf bytes.Buffer
	reqGw := gzip.NewWriter(&reqBuf)
	_, _ = reqGw.Write(originalRequestData)
	_ = reqGw.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		if !bytes.Equal(body, originalRequestData) {
			t.Errorf("decompressed request body doesn't match")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(originalResponseData)
	})

	middleware := GzipMiddleware(handler)

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(reqBuf.Bytes()))
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	contentEncoding := w.Header().Get("Content-Encoding")
	if contentEncoding != "gzip" {
		t.Errorf("expected Content-Encoding: gzip, got: %s", contentEncoding)
	}

	gr, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer func() { _ = gr.Close() }()

	decompressed, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	if !bytes.Equal(decompressed, originalResponseData) {
		t.Errorf("decompressed response doesn't match")
	}
}

func TestGzipMiddleware_ContentTypeCompression(t *testing.T) {
	testCases := []struct {
		name           string
		contentType    string
		shouldCompress bool
	}{
		{"JSON", "application/json", true},
		{"HTML", "text/html", true},
		{"Plain text", "text/plain", true},
		{"CSS", "text/css", true},
		{"JavaScript", "application/javascript", true},
		{"Image PNG", "image/png", false},
		{"Image JPEG", "image/jpeg", false},
		{"Binary", "application/octet-stream", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tc.contentType)
				_, _ = w.Write([]byte("test data"))
			})

			middleware := GzipMiddleware(handler)

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Accept-Encoding", "gzip")

			w := httptest.NewRecorder()
			middleware.ServeHTTP(w, req)

			contentEncoding := w.Header().Get("Content-Encoding")
			if tc.shouldCompress {
				if contentEncoding != "gzip" {
					t.Errorf("expected Content-Encoding: gzip for %s, got: %s", tc.contentType, contentEncoding)
				}
			} else {
				if contentEncoding == "gzip" {
					t.Errorf("should not compress %s, but got Content-Encoding: gzip", tc.contentType)
				}
			}
		})
	}
}

func TestGzipMiddleware_StatusCodeCompression(t *testing.T) {
	testCases := []struct {
		name           string
		statusCode     int
		shouldCompress bool
	}{
		{"OK", http.StatusOK, true},
		{"Created", http.StatusCreated, true},
		{"No Content", http.StatusNoContent, true},
		{"Bad Request", http.StatusBadRequest, false},
		{"Not Found", http.StatusNotFound, false},
		{"Internal Server Error", http.StatusInternalServerError, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte("test data"))
			})

			middleware := GzipMiddleware(handler)

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Accept-Encoding", "gzip")

			w := httptest.NewRecorder()
			middleware.ServeHTTP(w, req)

			contentEncoding := w.Header().Get("Content-Encoding")
			if tc.shouldCompress {
				if contentEncoding != "gzip" {
					t.Errorf("expected Content-Encoding: gzip for status %d, got: %s", tc.statusCode, contentEncoding)
				}
			} else {
				if contentEncoding == "gzip" {
					t.Errorf("should not compress status %d, but got Content-Encoding: gzip", tc.statusCode)
				}
			}

			if w.Code != tc.statusCode {
				t.Errorf("expected status code %d, got %d", tc.statusCode, w.Code)
			}
		})
	}
}
