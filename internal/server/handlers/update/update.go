// Package update provides HTTP handlers for updating individual metrics.
package update

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/lib/api/format"
)

// Updater interface for storage
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=Updater
type Updater interface {
	// UpdateGauge updates the gauge metric with the given key and value.
	// ctx: context for the operation.
	// key: the name of the gauge metric.
	// value: the value to update the gauge metric with.
	UpdateGauge(ctx context.Context, key string, value float64) error

	// UpdateCounter updates the counter metric with the given key and value.
	// ctx: context for the operation.
	// key: the name of the counter metric.
	// value: the value to update the counter metric with.
	UpdateCounter(ctx context.Context, key string, value int64) error

	// GetMetric retrieves the metric value for the given type and key.
	// ctx: context for the operation.
	// typ: the type of the metric (e.g., gauge, counter).
	// key: the name of the metric.
	// Returns the metric value as a string and an error if any.
	GetMetric(ctx context.Context, typ string, key string) (string, error)
}

// Backuper interface for backuper
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=Backuper
type Backuper interface {
	// SaveToStruct saves the metric data to a struct.
	// typ: the type of the metric (e.g., gauge, counter).
	// name: the name of the metric.
	// value: the value of the metric.
	SaveToStruct(typ string, name string, value string) error

	// SaveToFile saves the metric data to a file.
	SaveToFile()

	// IsSyncMode checks if the backup is in sync mode.
	// Returns true if the backup is in sync mode, false otherwise.
	IsSyncMode() bool
}

// New returns an HTTP handler function for updating metrics.
// log: the logger instance for logging.
// storage: the storage interface for updating metrics.
// backup: the backup interface for saving metrics.
func New(log *zap.Logger, storage Updater, backup Backuper) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.update.New"

		ctx := r.Context()
		log.With(
			zap.String("op", op),
			zap.String("request_id", middleware.GetReqID(ctx)),
		)

		name := chi.URLParam(r, "name")
		typ := chi.URLParam(r, "type")
		value := strings.TrimPrefix(r.URL.Path, "/update/"+typ+"/"+name+"/")

		if name == "" || value == "" {
			log.Error(
				"Name or Value is empty!",
				zap.String("name", name),
				zap.String("value", value))

			w.WriteHeader(http.StatusNotFound)
			return
		}

		databaseCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		switch typ {
		case format.Gauge:
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				log.Error("Failed to parse gauge value", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err = storage.UpdateGauge(databaseCtx, name, val)
			if err != nil {
				log.Error("Failed to update gauge value", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case format.Counter:
			val, err := strconv.ParseInt(value, 0, 64)
			if err != nil {
				log.Error("Failed to parse counter value", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err = storage.UpdateCounter(databaseCtx, name, val)
			if err != nil {
				log.Error("Failed to update counter value", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		default:
			log.Error("Undefined metric type", zap.String("type", typ))
			w.WriteHeader(http.StatusBadRequest)
			return

		}

		if backup.IsSyncMode() {
			backup.SaveToStruct(typ, name, value)
			backup.SaveToFile()
		}

		w.WriteHeader(http.StatusOK)
	}
}
