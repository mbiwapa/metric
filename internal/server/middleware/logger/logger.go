// Package logger provides middleware for logging HTTP requests and responses using zap.Logger.
package logger

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

// New returns a new logger middleware.
// This middleware logs the details of each HTTP request and response, including the method, path, request ID, status, bytes written, and duration.
// It uses the zap.Logger for structured logging.
//
// Parameters:
//   - log: A zap.Logger instance used for logging.
//
// Returns:
//   - A middleware function that wraps an http.Handler and logs the request and response details.
func New(log *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Add a component field to the logger for identifying the middleware.
		log = log.With(
			zap.String("component", "middleware/logger"),
		)

		log.Info("logger middleware enabled")

		// The actual middleware function that wraps the http.Handler.
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Create a new logger entry with request-specific fields.
			entry := log.With(
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("request_id", middleware.GetReqID(r.Context())),
			)
			// Wrap the response writer to capture the status and bytes written.
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				// Log the request completion details after the response is written.
				entry.Info("request completed",
					zap.Int("status", ww.Status()),
					zap.Int("bytes", ww.BytesWritten()),
					zap.String("duration", time.Since(t1).String()),
				)
			}()

			// Call the next handler in the chain.
			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
