package agent

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"sync"
)

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(nil, gzip.DefaultCompression)
		return w
	},
}

func Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer

	gz := gzipWriterPool.Get().(*gzip.Writer)
	defer func() {
		gz.Reset(io.Discard)
		gzipWriterPool.Put(gz)
	}()

	gz.Reset(&b)

	_, err := gz.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to write data to gzip: %v", err)
	}

	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %v", err)
	}

	return b.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer func() { _ = r.Close() }()

	var b bytes.Buffer
	_, err = b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read from gzip: %v", err)
	}
	return b.Bytes(), nil
}
