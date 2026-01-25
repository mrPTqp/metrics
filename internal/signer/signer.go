package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"go.uber.org/zap"
)

func Sign(content []byte, secretKey *string) (*string, error) {
	h, err := computeHMAC(content, secretKey)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func computeHMAC(content []byte, secretKey *string) (*string, error) {
	h := hmac.New(sha256.New, []byte(*secretKey))
	if _, err := h.Write(content); err != nil {
		return nil, err
	}
	hmac := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return &hmac, nil
}

func Verify(content []byte, hmac *string, secretKey *string, logger *zap.SugaredLogger) bool {
	if h, err := computeHMAC(content, secretKey); err != nil {
		logger.Errorw("failed to compute HMAC", err)
		return false
	} else {
		if hmac == nil || h == nil {
			logger.Warn("signature is missing")
			return false
		}
		if *hmac != *h {
			logger.Infof("signature mismatch:\nexpected: %s\nreceived: %s", *h, *hmac)
		}
		return *hmac == *h
	}
}
