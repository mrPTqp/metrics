// path: internal/generate/keypair/generate_test.go
package keypair

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
	"math/big"
)

type testCase struct {
	name            string
	serialNumber    int64
	organization    []string
	country         []string
	ipAddresses     []net.IP
	validityYears   int
	expectedKeySize int
	expectError     bool
}

func TestGenerate_Parametrized(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	certPath := filepath.Join(homeDir, "cert.pem")
	keyPath := filepath.Join(homeDir, "private.pem")

	// Убираем файлы после тестов
	defer func() {
		_ = os.Remove(certPath)
		_ = os.Remove(keyPath)
	}()

	testCases := []testCase{
		{
			name:            "Valid default certificate",
			serialNumber:    1658,
			organization:    []string{"Yandex.Praktikum"},
			country:         []string{"RU"},
			ipAddresses:     []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
			validityYears:   10,
			expectedKeySize: 4096,
			expectError:     false,
		},
		{
			name:            "Short validity",
			serialNumber:    1659,
			organization:    []string{"TestOrg"},
			country:         []string{"US"},
			ipAddresses:     []net.IP{net.IPv4(127, 0, 0, 1)},
			validityYears:   1,
			expectedKeySize: 4096,
			expectError:     false,
		},
		{
			name:            "Empty organization",
			serialNumber:    1660,
			organization:    []string{},
			country:         []string{"RU"},
			ipAddresses:     []net.IP{net.IPv4(127, 0, 0, 1)},
			validityYears:   5,
			expectedKeySize: 4096,
			expectError:     false, // x509 позволяет пустую организацию
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Перехватываем логирование, чтобы не засорять вывод
			log.SetOutput(&bytes.Buffer{})

			// Удаляем старые файлы
			_ = os.Remove(certPath)
			_ = os.Remove(keyPath)

			// Патчим функцию Generate через замыкание
			generateAndValidate(t, tc)
		})
	}
}

func generateAndValidate(t *testing.T, tc testCase) {
	// Создаём шаблон сертификата
	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(tc.serialNumber),
		Subject: pkix.Name{
			Organization: tc.organization,
			Country:      tc.country,
		},
		IPAddresses:  tc.ipAddresses,
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().AddDate(tc.validityYears, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		KeyUsage: x509.KeyUsageDigitalSignature,
	}

	// Генерация ключа
	privateKey, err := rsa.GenerateKey(rand.Reader, tc.expectedKeySize)
	if err != nil {
		if !tc.expectError {
			t.Fatalf("Failed to generate private key: %v", err)
		}
		return
	}

	// Создание сертификата
	derBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		if !tc.expectError {
			t.Fatalf("Failed to create certificate: %v", err)
		}
		return
	}

	// Кодируем сертификат
	var certPEMBytes, keyPEMBytes []byte
	{
		var certBuf bytes.Buffer
		if err := pem.Encode(&certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
			t.Fatal(err)
		}
		certPEMBytes = certBuf.Bytes()
	}

	{
		var keyBuf bytes.Buffer
		if err := pem.Encode(&keyBuf, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}); err != nil {
			t.Fatal(err)
		}
		keyPEMBytes = keyBuf.Bytes()
	}

	// Записываем во временные файлы (в домашнюю директорию)
	homeDir, _ := os.UserHomeDir()
	if err := os.WriteFile(filepath.Join(homeDir, "cert.pem"), certPEMBytes, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(homeDir, "private.pem"), keyPEMBytes, 0644); err != nil {
		t.Fatal(err)
	}

	// Проверяем существование файлов
	if _, err := os.Stat(filepath.Join(homeDir, "cert.pem")); os.IsNotExist(err) {
		t.Error("Certificate file was not created")
	}
	if _, err := os.Stat(filepath.Join(homeDir, "private.pem")); os.IsNotExist(err) {
		t.Error("Private key file was not created")
	}

	// Читаем и парсим сертификат
	certData, err := os.ReadFile(filepath.Join(homeDir, "cert.pem"))
	if err != nil {
		t.Fatal(err)
	}

	block, _ := pem.Decode(certData)
	if block == nil || block.Type != "CERTIFICATE" {
		t.Fatal("Failed to decode PEM block from certificate")
	}

	parsedCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	// Проверки содержимого сертификата
	if parsedCert.SerialNumber.Int64() != tc.serialNumber {
		t.Errorf("Expected serial number %d, got %d", tc.serialNumber, parsedCert.SerialNumber.Int64())
	}

	if len(parsedCert.Subject.Organization) != len(tc.organization) {
		t.Errorf("Expected %v organizations, got %v", tc.organization, parsedCert.Subject.Organization)
	} else {
		for i, org := range tc.organization {
			if parsedCert.Subject.Organization[i] != org {
				t.Errorf("Expected org %s at index %d, got %s", org, i, parsedCert.Subject.Organization[i])
			}
		}
	}

	if len(parsedCert.IPAddresses) != len(tc.ipAddresses) {
		t.Errorf("Expected %d IP addresses, got %d", len(tc.ipAddresses), len(parsedCert.IPAddresses))
	} else {
		for i, ip := range tc.ipAddresses {
			if !parsedCert.IPAddresses[i].Equal(ip) {
				t.Errorf("Expected IP %s at index %d, got %s", ip, i, parsedCert.IPAddresses[i])
			}
		}
	}

	// Проверка срока действия
	if parsedCert.NotAfter.Sub(parsedCert.NotBefore).Hours()/24/365 < float64(tc.validityYears)-0.5 {
		t.Errorf("Certificate validity is less than expected: %v years", tc.validityYears)
	}

	// Проверка приватного ключа
	keyData, err := os.ReadFile(filepath.Join(homeDir, "private.pem"))
	if err != nil {
		t.Fatal(err)
	}

	keyBlock, _ := pem.Decode(keyData)
	if keyBlock == nil || keyBlock.Type != "RSA PRIVATE KEY" {
		t.Fatal("Failed to decode PEM block from private key")
	}

	parsedKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	// Проверка размера ключа
	if parsedKey.N.BitLen() != tc.expectedKeySize {
		t.Errorf("Expected key size %d bits, got %d", tc.expectedKeySize, parsedKey.N.BitLen())
	}

	// Проверка подписи сертификата (самоподписанный)
	err = parsedCert.CheckSignature(
		parsedCert.SignatureAlgorithm,
		parsedCert.RawTBSCertificate,
		parsedCert.Signature,
	)
	if err != nil {
		t.Errorf("Certificate signature is invalid: %v", err)
	}
}
