package home

import (
	"net/http"
)

// AllMetricGeter interface for Metric repo
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=AllMetricGeter
type AllMetricGeter interface {
	GetAllMetrics() ([][]string, [][]string, error)
}

// New возвращает обработчик возвращающий HTML страницу со всеми доступными
func New(stor AllMetricGeter) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		gauge, counter, err := stor.GetAllMetrics()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

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
