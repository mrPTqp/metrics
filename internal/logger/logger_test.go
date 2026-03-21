package logger

import "testing"

func TestNewLogger_NotNilAndUsable(t *testing.T) {
	log := NewLogger()
	if log == nil {
		t.Fatal("NewLogger() returned nil")
	}


	log.Info("test log message")
	log.Sync() // ignore error, just ensure it can be called
}

