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
)

// Updater interface for storage
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=Updater
type Updater interface {
	UpdateBatch(ctx context.Context, gauges [][]string, counters [][]string) error
}

// Backuper interface for backuper
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=Backuper
type Backuper interface {
	SaveToStruct(typ string, name string, value string) error
	SaveToFile()
	IsSyncMode() bool
}

// NewJSON returned func for batch update
func NewJSON(log *zap.Logger, storage Updater, backup Backuper) http.HandlerFunc {

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
		enc := json.NewEncoder(w)
		if err := enc.Encode(metricsRequest); err != nil {
			log.Error("Error encoding response", zap.Error(err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

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
