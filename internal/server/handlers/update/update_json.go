package update

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/lib/api/format"
)

// NewJSON returned func for update
func NewJSON(log *zap.Logger, storage Updater, backup Backuper) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.update.NewJSON"

		log.With(
			zap.String("op", op),
			zap.String("request_id", middleware.GetReqID(r.Context())),
		)

		var metricRequest format.Metric

		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&metricRequest); err != nil {
			log.Error("Cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if metricRequest.ID == "" {
			log.Error("Name is empty!", zap.String("name", metricRequest.ID))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var updateErr error

		switch metricRequest.MType {
		case format.Gauge:
			updateErr = storage.UpdateGauge(metricRequest.ID, *metricRequest.Value)
		case format.Counter:
			updateErr = storage.UpdateCounter(metricRequest.ID, *metricRequest.Delta)
		default:
			log.Error("Undefined metric type", zap.String("type", metricRequest.MType))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if updateErr != nil {
			log.Error("Failed to update value", zap.Error(updateErr))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		if err := enc.Encode(metricRequest); err != nil {
			log.Error("Error encoding response", zap.Error(err))
			return
		}

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