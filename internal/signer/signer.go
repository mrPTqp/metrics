package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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

func Verify(content []byte, hmac *string, secretKey *string) bool {
	if h, err := computeHMAC(content, secretKey); err != nil {
		return false
	} else {
		return hmac == h
	}
}
