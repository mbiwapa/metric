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

// NewJSON возвращает обработчик для вывода метрики
func NewJSON(log *zap.Logger, storage MetricGeter, sha256key string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.value.NewJSON"

		ctx := r.Context()
		log.With(
			zap.String("op", op),
			zap.String("request_id", middleware.GetReqID(ctx)),
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

		databaseCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		value, err := storage.GetMetric(databaseCtx, metricRequest.MType, metricRequest.ID)
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
		body, err := json.Marshal(metricRequest)
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
