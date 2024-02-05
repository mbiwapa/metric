package value

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/lib/signature"
	storageErrors "github.com/mbiwapa/metric/internal/storage"
)

// MetricGeter interface for storage
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=MetricGeter
type MetricGeter interface {
	GetMetric(ctx context.Context, typ string, key string) (string, error)
}

// New возвращает обработчик для вывода метрики
func New(log *zap.Logger, storage MetricGeter, sha256key string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.value.New"

		ctx := r.Context()
		log.With(
			zap.String("op", op),
			zap.String("request_id", middleware.GetReqID(ctx)),
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

		databaseCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		value, err := storage.GetMetric(databaseCtx, typ, name)
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

		if sha256key != "" {
			hashStr := signature.GetHash(sha256key, string(value), log)
			w.Header().Set("HashSHA256", hashStr)
		}

		w.Write([]byte(value))
	}
}
