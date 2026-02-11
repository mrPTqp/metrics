package middleware

import (
	"net/http"
	"strings"

	"go.uber.org/zap"
	"github.com/mrPTqp/metrics/internal/contextkey"
)

func GzipMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log := contextkey.LoggerFromContext(r.Context())

        var (
            cw   *compressWriter
            ow   http.ResponseWriter = w
        )

        acceptEncoding := r.Header.Get("Accept-Encoding")
        supportsGzip := strings.Contains(acceptEncoding, "gzip")

        if supportsGzip {
            cw = newCompressWriter(w, supportsGzip, log)
            ow = cw
        }

        defer func() {
            if cw != nil {
                if closeErr := cw.Close(); closeErr != nil {
                    log.Error("gzip writer close error", zap.Error(closeErr))
                }
            }
        }()

        contentEncoding := r.Header.Get("Content-Encoding")
        sendsGzip := strings.Contains(contentEncoding, "gzip")
        if sendsGzip {
            cr, err := newCompressReader(r.Body)
            if err != nil {
                log.Error("Error creating compress reader", zap.Error(err))
                w.WriteHeader(http.StatusInternalServerError)
                return
            }
            r.Body = cr
            defer func() {
                _ = cr.Close()
            }()
        }

        next.ServeHTTP(ow, r)
    })
}
