package value

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	storageErrors "github.com/mbiwapa/metric/internal/storage"
)

// MetricGeter interface for storage
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=MetricGeter
type MetricGeter interface {
	GetMetric(typ string, key string) (string, error)
}

// New возвращает обработчик для вывода метрики
func New(log *zap.Logger, storage MetricGeter) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.value.New"
		log.With(
			zap.String("op", op),
			zap.String("request_id", middleware.GetReqID(r.Context())),
		)
		typ := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")

		if name == "" || typ == "" {
			log.Error(
				"Name or Type is empty!",
				zap.String("name", name),
				zap.String("type", typ))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		value, err := storage.GetMetric(typ, name)
		if errors.Is(err, storageErrors.ErrMetricNotFound) {
			log.Info(
				"Metric is not found",
				zap.String("name", name),
				zap.String("type", typ))
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil && !errors.Is(err, storageErrors.ErrMetricNotFound) {
			log.Error("Failed to get metric", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write([]byte(value))
	}
}
