// Package home provides HTTP handlers for serving an HTML page with all available metrics.
package home

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/lib/signature"
)

// AllMetricGeter defines the methods required to retrieve all metrics from the storage.
// It is used to abstract the data retrieval logic, allowing for different implementations.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=AllMetricGeter
type AllMetricGeter interface {
	// GetAllMetrics retrieves all gauge and counter metrics from the storage.
	//
	// Parameters:
	//   - ctx: A context.Context instance for managing request-scoped values, cancellation, and deadlines.
	//
	// Returns:
	//   - [][]string: A slice of slices containing gauge metrics, where each inner slice represents a metric with its name and value.
	//   - [][]string: A slice of slices containing counter metrics, where each inner slice represents a metric with its name and value.
	//   - error: An error object if there is an issue retrieving the metrics, otherwise nil.
	GetAllMetrics(ctx context.Context) ([][]string, [][]string, error)
}

// New returns an HTTP handler function that serves an HTML page with all available metrics.
// It logs the request, retrieves metrics from the storage, and constructs an HTML response.
// If a SHA256 key is provided, it also includes a hash of the response body in the headers.
//
// Parameters:
//   - log: A zap.Logger instance for logging.
//   - storage: An implementation of the AllMetricGeter interface to retrieve metrics.
//   - sha256key: A string key used to generate a SHA256 hash of the response body.
//
// Returns:
//   - An http.HandlerFunc that handles HTTP requests and serves the metrics page.
func New(log *zap.Logger, storage AllMetricGeter, sha256key string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.home.New"

		ctx := r.Context()
		log = log.With(
			zap.String("op", op),
			zap.String("request_id", middleware.GetReqID(ctx)),
		)

		databaseCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		gauge, counter, err := storage.GetAllMetrics(databaseCtx)
		if err != nil {
			log.Error("Failed to get all metrics", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Info("Metrics received", zap.Any("gauge", gauge), zap.Any("counter", counter))

		body := "<!DOCTYPE html><html><head><title>Метрики</title><body><h1>Метрики</h1><ul>"

		if len(gauge) > 0 {
			for _, metric := range gauge {
				body += "<li>" + metric[0] + ": " + metric[1] + "</li>"
			}
		}
		if len(counter) > 0 {
			for _, metric := range counter {
				body += "<li>" + metric[0] + ": " + metric[1] + "</li>"
			}
		}

		body += "</ul></body></html>"
		w.Header().Set("Content-Type", "text/html")

		if sha256key != "" {
			hashStr := signature.GetHash(sha256key, body, log)
			w.Header().Set("HashSHA256", hashStr)
		}

		w.Write([]byte(body))
	}
}
