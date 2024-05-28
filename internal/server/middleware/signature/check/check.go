package check

import (
	"bytes"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/lib/signature"
)

// New creates a middleware function that checks the signature of incoming HTTP requests.
// It verifies that the SHA256 hash of the request body matches the expected hash derived from a secret key.
// If the hashes do not match, it responds with a 400 Bad Request status code.
//
// Parameters:
// - key: A secret key used to generate the expected hash of the request body.
// - log: A zap.Logger instance for logging information and errors.
//
// Returns:
// - A middleware function that can be used with an HTTP handler to check request signatures.
func New(key string, log *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log = log.With(
			zap.String("op", "middleware.signature.Check"),
		)

		fn := func(w http.ResponseWriter, r *http.Request) {
			sha256Hash := r.Header.Get("HashSHA256")

			if sha256Hash != "" && key != "" {
				log.Info("Keys", zap.String("responseHash", sha256Hash))

				body, err := io.ReadAll(r.Body)
				if err != nil {
					log.Error("Cannot read body", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				// Restore the request body for further processing
				r.Body = io.NopCloser(bytes.NewBuffer(body))

				hashStr := signature.GetHash(key, string(body), log)
				if hashStr != sha256Hash {
					log.Error("Signature mismatch", zap.String("hashRequest", sha256Hash))
					w.WriteHeader(http.StatusBadRequest)
					return
				}

			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
