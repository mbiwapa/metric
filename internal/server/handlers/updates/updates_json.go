package updates

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/lib/api/format"
	"github.com/mbiwapa/metric/internal/lib/signature"
)

// Updater interface for storage
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=Updater
type Updater interface {
	// UpdateBatch updates a batch of gauge and counter metrics in the storage.
	// It takes a context for cancellation, and slices of gauge and counter metrics.
	UpdateBatch(ctx context.Context, gauges [][]string, counters [][]string) error
}

// Backuper interface for backuper
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=Backuper
type Backuper interface {
	// SaveToStruct saves a metric to a backup structure.
	// It takes the type, name, and value of the metric.
	SaveToStruct(typ string, name string, value string) error

	// SaveToFile saves the backup structure to a file.
	SaveToFile()

	// IsSyncMode checks if the backup is in synchronous mode.
	IsSyncMode() bool
}

// NewJSON returns an HTTP handler function for batch updating metrics.
// It takes a logger, storage updater, backup handler, and an optional SHA256 key for response hashing.
func NewJSON(log *zap.Logger, storage Updater, backup Backuper, sha256key string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.updates.NewJSON"

		ctx := r.Context()
		log.With(
			zap.String("op", op),
			zap.String("request_id", middleware.GetReqID(ctx)),
		)

		var metricsRequest []format.Metric

		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&metricsRequest); err != nil {
			log.Error("Cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var gauges [][]string
		var counters [][]string
		for _, metric := range metricsRequest {
			switch metric.MType {
			case format.Gauge:
				gauges = append(gauges, []string{metric.ID, strconv.FormatFloat(*metric.Value, 'f', -1, 64)})
			case format.Counter:
				counters = append(counters, []string{metric.ID, strconv.FormatInt(*metric.Delta, 10)})
			default:
				log.Error("Unknown metric type", zap.String("type", metric.MType))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			err := backupHandler(log, backup, metric)
			if err != nil {
				log.Error("Cannot backup metric", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		databaseCtx, cancel := context.WithTimeout(ctx, 11*time.Second)
		defer cancel()
		err := storage.UpdateBatch(databaseCtx, gauges, counters)

		if err != nil {
			log.Error("Failed to batch update", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		body, err := json.Marshal(metricsRequest)
		if err != nil {
			log.Error("Error encoding response", zap.Error(err))
			return
		}

		if sha256key != "" {
			hashStr := signature.GetHash(sha256key, string(body), log)
			w.Header().Set("HashSHA256", hashStr)
		}

		w.Write(body)
		w.WriteHeader(http.StatusOK)
	}
}

// backupHandler handles the backup of a single metric.
// It takes a logger, backup handler, and the metric to be backed up.
// Returns an error if the metric type is undefined or if the backup fails.
func backupHandler(log *zap.Logger, backup Backuper, metric format.Metric) error {
	const op = "handlers.updates.backup"
	if backup.IsSyncMode() {
		var backupVal string
		switch metric.MType {
		case format.Gauge:
			backupVal = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		case format.Counter:
			backupVal = strconv.FormatInt(*metric.Delta, 10)
		default:
			log.Error("Undefined metric type", zap.String("type", metric.MType))
			return fmt.Errorf("%s: %s", op, "Undefined metric type")
		}
		backup.SaveToStruct(metric.MType, metric.ID, backupVal)
		backup.SaveToFile()
	}
	return nil
}
