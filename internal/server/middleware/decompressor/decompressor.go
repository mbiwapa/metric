package decompressor

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// New returns a middleware function that compresses the response body if the
// request contains the "Content-Encoding: gzip" header.
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

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

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

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
