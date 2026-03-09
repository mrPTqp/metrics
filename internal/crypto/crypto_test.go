package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"testing"
)

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	cert := &x509.Certificate{
		PublicKey: &privateKey.PublicKey,
	}

	plain := []byte("secret data payload")

	encrypted, err := Encrypt(cert, plain)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if len(encrypted) == 0 {
		t.Fatal("Encrypt() returned empty payload")
	}

	decrypted, err := Decrypt(privateKey, encrypted)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if string(decrypted) != string(plain) {
		t.Errorf("Decrypt() = %q, want %q", string(decrypted), string(plain))
	}
}

func TestDecrypt_InvalidShortPayload(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	payload := []byte("too short")
	if _, err := Decrypt(privateKey, payload); err == nil {
		t.Fatal("Decrypt() error = nil, want non-nil for short payload")
	}
}

func TestReadCertificate_FileNotFound(t *testing.T) {
	if _, err := ReadCertificate("non-existent-cert.pem"); err == nil {
		t.Fatal("ReadCertificate() error = nil, want non-nil")
	}
}

func TestReadPrivateKey_FileNotFound(t *testing.T) {
	if _, err := ReadPrivateKey("non-existent-key.pem"); err == nil {
		t.Fatal("ReadPrivateKey() error = nil, want non-nil")
	}
}

func TestEncrypt_InvalidPublicKeyType(t *testing.T) {
	// Certificate without RSA public key should cause an error
	cert := &x509.Certificate{
		PublicKey: struct{}{},
	}
	_, err := Encrypt(cert, []byte("data"))
	if err == nil {
		t.Fatal("Encrypt() error = nil, want non-nil for invalid public key type")
	}
	if !errors.Is(err, errors.New("public key read error")) {
		if err.Error() == "" {
			t.Fatalf("Encrypt() returned unexpected error: %v", err)
		}
	}
}

