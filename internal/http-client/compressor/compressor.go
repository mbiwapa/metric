package compressor

import (
	"bytes"
	"compress/gzip"
	"io"

	"go.uber.org/zap"
)

// Compressor struct for compressing data
type Compressor struct {
	Logger *zap.Logger
}

// New creates a new Compressor instance
func New(log *zap.Logger) *Compressor {
	return &Compressor{
		Logger: log,
	}
}

// GetCompressedData returns a compressed reader for the given data.
func (compressor Compressor) GetCompressedData(data []byte) (io.Reader, error) {
	b := new(bytes.Buffer)
	w, err := gzip.NewWriterLevel(b, gzip.BestSpeed)
	if err != nil {
		compressor.Logger.Error("error init gzip writer")
		return nil, err
	}
	_, err = w.Write(data)
	if err != nil {
		compressor.Logger.Error("error compressing data")
		return nil, err
	}
	err = w.Close()
	w.Reset(b)
	if err != nil {
		compressor.Logger.Error("error closing writer")
		return nil, err
	}

	return b, nil
}
