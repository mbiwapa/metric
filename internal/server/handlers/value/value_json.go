package value

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/lib/api/format"
	storageErrors "github.com/mbiwapa/metric/internal/storage"
)

// NewJSON возвращает обработчик для вывода метрики
func NewJSON(log *zap.Logger, storage MetricGeter) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.value.NewJSON"
		log.With(
			zap.String("op", op),
			zap.String("request_id", middleware.GetReqID(r.Context())),
		)

		var metricRequest format.Metric

		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&metricRequest); err != nil {
			log.Error(
				"Cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if metricRequest.ID == "" || metricRequest.MType == "" {
			log.Error(
				"Name or Type is empty!",
				zap.String("name", metricRequest.ID),
				zap.String("type", metricRequest.MType))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		value, err := storage.GetMetric(metricRequest.MType, metricRequest.ID)
		if errors.Is(err, storageErrors.ErrMetricNotFound) {
			log.Info(
				"Metric is not found",
				zap.String("name", metricRequest.ID),
				zap.String("type", metricRequest.MType))
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil && !errors.Is(err, storageErrors.ErrMetricNotFound) {
			log.Error("Failed to get metric", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

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
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		if err := enc.Encode(metricRequest); err != nil {
			log.Error("Error encoding response", zap.Error(err))
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
