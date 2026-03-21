package signer

import (
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestSignAndVerify(t *testing.T) {
	secret := "super-secret"
	content := []byte("important payload")
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name      string
		signature func() *string
		wantOK    bool
	}{
		{
			name: "valid signature",
			signature: func() *string {
				sig, _ := Sign(content, &secret)
				if sig == nil || *sig == "" {
					t.Fatalf("Sign() returned empty signature")
				}
				return sig
			},
			wantOK: true,
		},
		{
			name: "mismatched signature",
			signature: func() *string {
				badSig := "invalid-signature"
				return &badSig
			},
			wantOK: false,
		},
		{
			name: "missing signature",
			signature: func() *string {
				return nil
			},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := tt.signature()
			if ok := Verify(content, sig, &secret, logger); ok != tt.wantOK {
				t.Errorf("Verify() = %v, want %v", ok, tt.wantOK)
			}
		})
	}
}


