package check

import (
	"io"
	"net/http"

	"github.com/mbiwapa/metric/internal/lib/signature"
	"go.uber.org/zap"
)

// New function check signature of request
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
					log.Error("Can't read request body", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

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
