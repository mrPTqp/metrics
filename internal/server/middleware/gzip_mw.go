package middleware

import (
	"net/http"
	"strings"

	"github.com/mrPTqp/metrics/internal/contextkey"
	"go.uber.org/zap"
)

// Middleware для распаковки запросов с gzip сжатием и сжатия ответов.
// По заголовку "Accept-Encoding" определяет необходимость распаковки ответа.
// По заголовку "Content-Encoding" определяет необходимость сжатия ответа
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := contextkey.LoggerFromContext(r.Context())

		var (
			cw *compressWriter
			ow = w
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
