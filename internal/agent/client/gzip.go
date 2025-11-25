package agent

import (
	"bytes"
	"compress/gzip"
	"fmt"
)

func Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err := gz.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to gzip: %v", err)
	}
	err = gz.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %v", err)
	}
	return b.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer r.Close()

	var b bytes.Buffer
	_, err = b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read from gzip: %v", err)
	}
	return b.Bytes(), nil
}
