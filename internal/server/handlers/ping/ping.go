package ping

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

// Pinger interface for Metric repo
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=AllMetricGeter
type Pinger interface {
	Ping(ctx context.Context) error
}

// New возвращает обработчик проверяющий бд на доступность
func New(log *zap.Logger, storage Pinger) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.ping.New"

		ctx := r.Context()

		log.With(
			zap.String("op", op),
			zap.String("request_id", middleware.GetReqID(ctx)),
		)

		databaseCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		err := storage.Ping(databaseCtx)
		if err != nil {
			log.Error("Database is unvailable", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Info("Database is available")
		w.WriteHeader(http.StatusOK)
	}
}
