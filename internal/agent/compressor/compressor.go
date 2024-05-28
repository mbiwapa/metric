// Package compressor provides a compressor that compresses data using the gzip algorithm.
// It also provides a function to create a new Compressor instance and a function to get a compressed reader for the given data.
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
// log: A zap.Logger instance used for logging errors and information.
// Returns a pointer to a new Compressor instance.
func New(log *zap.Logger) *Compressor {
	return &Compressor{
		Logger: log,
	}
}

// GetCompressedData returns a compressed reader for the given data.
// data: A byte slice containing the data to be compressed.
// Returns an io.Reader containing the compressed data and an error if any occurred during compression.
func (compressor Compressor) GetCompressedData(data []byte) (io.Reader, error) {
	b := new(bytes.Buffer)
	w, err := gzip.NewWriterLevel(b, gzip.BestSpeed)
	if err != nil {
		compressor.Logger.Error("error init gzip writer", zap.Error(err))
		return nil, err
	}
	_, err = w.Write(data)
	if err != nil {
		compressor.Logger.Error("error compressing data", zap.Error(err))
		return nil, err
	}
	err = w.Close()
	w.Reset(b)
	if err != nil {
		compressor.Logger.Error("error closing writer", zap.Error(err))
		return nil, err
	}

	return b, nil
}
