package middleware

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func writeJSONError(w http.ResponseWriter, message string, statusCode int, logger *zap.Logger) {
	logger.Warn("Sending JSON error response",
		zap.Int("status", statusCode),
		zap.String("error", message))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}