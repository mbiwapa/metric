package home

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

// AllMetricGeter interface for Metric repo
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=AllMetricGeter
type AllMetricGeter interface {
	GetAllMetrics() ([][]string, [][]string, error)
}

// New возвращает обработчик возвращающий HTML страницу со всеми доступными
func New(log *zap.Logger, storage AllMetricGeter) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.home.New"

		log.With(
			zap.String("op", op),
			zap.String("request_id", middleware.GetReqID(r.Context())),
		)

		gauge, counter, err := storage.GetAllMetrics()
		if err != nil {
			log.Error("Failed to get all metrics", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Info("Metrics received", zap.Any("gauge", gauge), zap.Any("counter", counter))

		body := "<!DOCTYPE html><html><head><title>Метрики</title><body><h1>Метрики</h1><ul>"

		if len(gauge) > 0 {
			for _, metric := range gauge {
				body += "<li>" + metric[0] + ": " + metric[1] + "</li>"
			}
		}
		if len(counter) > 0 {
			for _, metric := range counter {
				body += "<li>" + metric[0] + ": " + metric[1] + "</li>"
			}
		}

		body += "</ul></body></html>"

		w.Write([]byte(body))
	}
}
