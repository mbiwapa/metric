package update

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/lib/api/format"
	"github.com/mbiwapa/metric/internal/lib/signature"
)

// NewJSON returns an HTTP handler function for updating metrics.
// It handles JSON requests, updates the metric in the storage, and optionally performs a backup.
//
// Parameters:
// - log: A zap.Logger instance for logging.
// - storage: An Updater interface for updating metrics in the storage.
// - backup: A Backuper interface for performing backups.
// - sha256key: A string key used for generating SHA256 hash.
//
// Returns:
// - An http.HandlerFunc that processes the update request.
func NewJSON(log *zap.Logger, storage Updater, backup Backuper, sha256key string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.update.NewJSON"

		ctx := r.Context()
		log.With(
			zap.String("op", op),
			zap.String("request_id", middleware.GetReqID(ctx)),
		)

		var metricRequest format.Metric

		// Decode the JSON request body into metricRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&metricRequest); err != nil {
			log.Error("Cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Check if the metric ID is empty
		if metricRequest.ID == "" {
			log.Error("Name is empty!", zap.String("name", metricRequest.ID))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Create a context with a timeout for database operations
		databaseCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		var updateErr error

		// Update the metric based on its type
		switch metricRequest.MType {
		case format.Gauge:
			updateErr = storage.UpdateGauge(databaseCtx, metricRequest.ID, *metricRequest.Value)
		case format.Counter:
			updateErr = storage.UpdateCounter(databaseCtx, metricRequest.ID, *metricRequest.Delta)
			databaseGetCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			stringVal, err := storage.GetMetric(databaseGetCtx, metricRequest.MType, metricRequest.ID)
			if err != nil {
				log.Error("Failed to get metric", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			newVal, err := strconv.ParseInt(stringVal, 0, 64)
			if err != nil {
				log.Error("Failed to parse int", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			metricRequest.Delta = &newVal
		default:
			log.Error("Undefined metric type", zap.String("type", metricRequest.MType))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Handle update errors
		if updateErr != nil {
			log.Error("Failed to update value", zap.Error(updateErr))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Set response content type to JSON
		w.Header().Set("Content-Type", "application/json")
		body, err := json.Marshal(metricRequest)
		if err != nil {
			log.Error("Error encoding response", zap.Error(err))
			return
		}

		// Generate and set SHA256 hash if key is provided
		if sha256key != "" {
			hashStr := signature.GetHash(sha256key, string(body), log)
			w.Header().Set("HashSHA256", hashStr)
		}

		// Write the response body
		w.Write(body)

		// Perform backup if in sync mode
		if backup.IsSyncMode() {
			var backupVal string
			switch metricRequest.MType {
			case format.Gauge:
				backupVal = strconv.FormatFloat(*metricRequest.Value, 'f', -1, 64)
			case format.Counter:
				backupVal = strconv.FormatInt(*metricRequest.Delta, 10)
			default:
				log.Error("Undefined metric type", zap.String("type", metricRequest.MType))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			backup.SaveToStruct(metricRequest.MType, metricRequest.ID, backupVal)
			backup.SaveToFile()
		}
		w.WriteHeader(http.StatusOK)
	}
}
