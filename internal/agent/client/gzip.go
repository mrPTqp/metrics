package agent

import (
	"bytes"
	"compress/gzip"
	"fmt"
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
	gz.Reset(&b)

	_, err := gz.Write(data)
	if err != nil {
		gz.Close() 
		gzipWriterPool.Put(gz)
		return nil, fmt.Errorf("failed write data to gzip: %v", err)
	}

	err = gz.Close()
	if err != nil {
		gzipWriterPool.Put(gz)
		return nil, fmt.Errorf("failed to close gzip writer: %v", err)
	}

	compressed := b.Bytes()
	gzipWriterPool.Put(gz)

	return compressed, nil
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
