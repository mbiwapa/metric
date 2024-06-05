// Package decompressor provides middleware for decompressing HTTP request bodies
// that are compressed using gzip.
package decompressor

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// New returns a middleware function that decompresses the request body if the
// request contains the "Content-Encoding: gzip" header.
//
// Parameters:
// - log: A zap.Logger instance for logging.
//
// Returns:
// - A middleware function that can be used with an HTTP handler.
func New(log *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log = log.With(
			zap.String("component", "middleware/decompressor"),
		)

		log.Info("decompressor middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")

			if sendsGzip {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					log.Error("failed init decompressor", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer func(cr *compressReader) {

					err := cr.Close()
					if err != nil {
						log.Error("failed closing compress reader", zap.Error(err))
					}
				}(cr)
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// compressReader wraps an io.ReadCloser and a gzip.Reader to provide
// decompression functionality.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// newCompressReader creates a new compressReader instance.
//
// Parameters:
// - r: An io.ReadCloser representing the compressed data.
//
// Returns:
// - A pointer to a compressReader instance.
// - An error if the gzip.Reader could not be created.
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read reads decompressed data into p.
//
// Parameters:
// - p: A byte slice to store the decompressed data.
//
// Returns:
// - The number of bytes read.
// - An error if there was an issue during reading.
func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close closes both the underlying io.ReadCloser and the gzip.Reader.
//
// Returns:
// - An error if there was an issue during closing.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
