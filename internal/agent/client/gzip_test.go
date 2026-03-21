package agent

import "testing"

func TestCompressDecompress_RoundTrip(t *testing.T) {
	original := []byte("some test data for gzip")

	compressed, err := Compress(original)
	if err != nil {
		t.Fatalf("Compress() error = %v, want nil", err)
	}
	if len(compressed) == 0 {
		t.Fatal("Compress() returned empty slice")
	}

	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress() error = %v, want nil", err)
	}
	if string(decompressed) != string(original) {
		t.Errorf("Decompress() = %q, want %q", string(decompressed), string(original))
	}
}

func TestDecompress_InvalidData(t *testing.T) {
	invalid := []byte("not a valid gzip stream")

	if _, err := Decompress(invalid); err == nil {
		t.Fatal("Decompress() error = nil, want non-nil for invalid data")
	}
}

