package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

var (
	ErrAESCipherCreate = errors.New("error create aes cipher")
	ErrGCMCreate       = errors.New("error create gcm")
)

func ReadCertificate(path string) (*x509.Certificate, error) {
	certificateBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	certificatePemBlock, _ := pem.Decode(certificateBytes)
	if certificatePemBlock == nil {
		return nil, fmt.Errorf("certificate not found")
	}

	certificate, err := x509.ParseCertificate(certificatePemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return certificate, nil
}

func ReadPrivateKey(path string) (*rsa.PrivateKey, error) {
	privateKeyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	privateKeyPemBlock, _ := pem.Decode(privateKeyBytes)
	if privateKeyPemBlock == nil {
		return nil, fmt.Errorf("private key not found")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyPemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func Encrypt(cert *x509.Certificate, body []byte) ([]byte, error) {
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("generate aes key: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("%w", ErrAESCipherCreate)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w", ErrGCMCreate)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key read error")
	}

	ciphertext := gcm.Seal(nil, nonce, body, nil)	
	
	encryptedKey, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, aesKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt aes key: %w", err)
	}

	var buf bytes.Buffer
	buf.Write(encryptedKey)
	buf.Write(nonce)
	buf.Write(ciphertext)
	
	return buf.Bytes(), nil
}

func Decrypt(privateKey *rsa.PrivateKey, payload []byte) ([]byte, error) {
	rsaBlockSize := privateKey.Size()

	if len(payload) < rsaBlockSize {
		return nil, fmt.Errorf("payload too short")
	}

	encryptedKey := payload[:rsaBlockSize]
	rest := payload[rsaBlockSize:]

	aesKey, err := rsa.DecryptPKCS1v15(nil, privateKey, encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting aes key: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("%w", ErrAESCipherCreate)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w", ErrGCMCreate)
	}

	nonceSize := gcm.NonceSize()
	if len(rest) < nonceSize {
		return nil, fmt.Errorf("payload too short")
	}

	nonce := rest[:nonceSize]
	ciphertext := rest[nonceSize:]

	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("error decrypting payload: %w", err)
	}

	return decrypted, nil
}
