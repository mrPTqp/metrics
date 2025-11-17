package middleware

import (
	"net/http"
	"strings"

	"go.uber.org/zap"
)

func GzipMiddleware(h http.HandlerFunc, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		logger.Debugf("accept encoding value %s", acceptEncoding)
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		logger.Debugf("supports gzip %b", supportsGzip)
		if supportsGzip {
			logger.Debug("try compress")
			cw := newCompressWriter(w, supportsGzip, logger)
			ow = cw
			defer func() {
				cw.Close()
			}()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				logger.Error("error creating compress reader", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer func() {
				_ = cr.Close()
			}()
		}

		h.ServeHTTP(ow, r)
	}
}
