package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"go.uber.org/zap"
)

func Sign(content []byte, secretKey string) (string, error) {
	h, err := computeHMAC(content, secretKey)
	if err != nil {
		return "", err
	}
	return h, nil
}

func computeHMAC(content []byte, secretKey string) (string, error) {
	h := hmac.New(sha256.New, []byte(secretKey))
	if _, err := h.Write(content); err != nil {
		return "", err
	}
	hmac := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return hmac, nil
}

func Verify(content []byte, receivedHMAC string, secretKey string, logger *zap.SugaredLogger) bool {
	expectedHMAC, err := computeHMAC(content, secretKey)
	if err != nil {
		logger.Errorw("failed to compute HMAC", "error", err)
		return false
	}

	if receivedHMAC == "" || expectedHMAC == "" {
		logger.Warn("signature is missing")
		return false
	}

	receivedMAC, err := base64.StdEncoding.DecodeString(receivedHMAC)
	if err != nil {
		logger.Warnw("failed to decode received HMAC", "error", err)
		return false
	}

	expectedMAC, err := base64.StdEncoding.DecodeString(expectedHMAC)
	if err != nil {
		logger.Errorw("failed to decode expected HMAC", "error", err)
		return false
	}

	if hmac.Equal(receivedMAC, expectedMAC) {
		logger.Info("signature is valid")
		return true
	}

	logger.Errorf("signature mismatch:\nexpected: %s\nreceived: %s", expectedHMAC, receivedHMAC)
	return false
}
