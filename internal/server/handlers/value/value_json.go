// Package value provides HTTP handlers for processing metric requests and responding with metric data.
package value

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/lib/api/format"
	"github.com/mbiwapa/metric/internal/lib/signature"
	storageErrors "github.com/mbiwapa/metric/internal/storage"
)

// NewJSON returns an HTTP handler function that processes metric requests and responds with the metric data in JSON format.
// It logs the request, decodes the JSON body, retrieves the metric from storage, and writes the response.
//
// Parameters:
// - log: A zap.Logger instance for logging.
// - storage: An implementation of the MetricGeter interface for retrieving metrics from storage.
// - sha256key: A string key used for generating SHA256 hash of the response body.
//
// Returns:
// - An http.HandlerFunc that handles the HTTP request and response.
func NewJSON(log *zap.Logger, storage MetricGeter, sha256key string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.value.NewJSON"

		ctx := r.Context()
		log.With(
			zap.String("op", op),
			zap.String("request_id", middleware.GetReqID(ctx)),
		)

		var metricRequest format.Metric

		// Decode the JSON request body into metricRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&metricRequest); err != nil {
			log.Error(
				"Cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Validate the metric request
		if metricRequest.ID == "" || metricRequest.MType == "" {
			log.Error(
				"Name or Type is empty!",
				zap.String("name", metricRequest.ID),
				zap.String("type", metricRequest.MType))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Create a context with a timeout for the database operation
		databaseCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		// Retrieve the metric value from storage
		value, errStor := storage.GetMetric(databaseCtx, metricRequest.MType, metricRequest.ID)
		if errors.Is(errStor, storageErrors.ErrMetricNotFound) {
			log.Info(
				"Metric is not found",
				zap.String("name", metricRequest.ID),
				zap.String("type", metricRequest.MType))
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errStor != nil && !errors.Is(errStor, storageErrors.ErrMetricNotFound) {
			log.Error("Failed to get metric", zap.Error(errStor))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Parse the metric value based on its type
		switch metricRequest.MType {
		case format.Gauge:
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				log.Error("Failed to parse gauge value", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			metricRequest.Value = &val
		case format.Counter:
			val, err := strconv.ParseInt(value, 0, 64)
			if err != nil {
				log.Error("Failed to parse counter value", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			metricRequest.Delta = &val
		default:
		}

		// Set the response content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Marshal the metric request into JSON
		body, err := json.Marshal(metricRequest)
		if err != nil {
			log.Error("Error encoding response", zap.Error(err))
			return
		}

		// If a SHA256 key is provided, generate and set the hash header
		if sha256key != "" {
			hashStr := signature.GetHash(sha256key, string(body), log)
			w.Header().Set("HashSHA256", hashStr)
		}

		// Write the response body and set the status code to OK
		w.Write(body)
		w.WriteHeader(http.StatusOK)
	}
}
