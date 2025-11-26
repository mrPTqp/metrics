package middleware

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"

    "github.com/mrPTqp/metrics/internal/signer"
    "go.uber.org/zap"
)

func writeJSONError(w http.ResponseWriter, message string, statusCode int, logger *zap.SugaredLogger) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    _ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func SignMiddleware(h http.HandlerFunc, key string, logger *zap.SugaredLogger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var bodyContent []byte
        var err error

        if r.Body != nil {
            bodyContent, err = io.ReadAll(r.Body)
            if err != nil {
                logger.Errorw("failed to read request body", "error", err)
                writeJSONError(w, "invalid request body", http.StatusBadRequest, logger)
                return
            }
            r.Body = io.NopCloser(bytes.NewReader(bodyContent))
        }

        if len(bodyContent) > 0 {
            signHeader := r.Header.Get("HashSHA256")
            if signHeader == "" {
                logger.Warn("missing HashSHA256 header")
                writeJSONError(w, "missing signature", http.StatusBadRequest, logger)
                return
            }

            if !signer.Verify(bodyContent, &signHeader, &key, logger) {
                logger.Warn("invalid signature in request")
                writeJSONError(w, "invalid signature", http.StatusBadRequest, logger)
                return
            }
        }

        ww := &responseWriter{ResponseWriter: w, body: &bytes.Buffer{}}
        h.ServeHTTP(ww, r)

        if ww.body.Len() > 0 {
            responseBody := ww.body.Bytes()
            sign, err := signer.Sign(responseBody, &key)
            if err != nil {
                logger.Errorw("failed to sign response", "error", err)
                writeJSONError(w, "failed to sign response", http.StatusInternalServerError, logger)
                return
            }
            ww.Header().Set("HashSHA256", *sign)
        }

        for k, v := range ww.Header() {
            w.Header()[k] = v
        }
        w.WriteHeader(ww.code)
        _, _ = w.Write(ww.body.Bytes())
    }
}

type responseWriter struct {
    http.ResponseWriter
    body *bytes.Buffer
    code int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.code = code
    rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
    return rw.body.Write(b)
}
