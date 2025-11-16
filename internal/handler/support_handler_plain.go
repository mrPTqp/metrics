package handler

import (
	"net/http"
)

func (mh *MetricHandler) DBHealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if !mh.service.Ping() {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}
