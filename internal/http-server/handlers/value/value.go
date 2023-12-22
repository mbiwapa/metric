package value

import (
	"net/http"

	"github.com/go-chi/chi"
)

// MetricGeter interface for storage
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=MetricGeter
type MetricGeter interface {
	GetMetric(typ string, key string) (string, error)
}

// New возвращает обработчик для вывода метрики
func New(stor MetricGeter) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		typ := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")

		if name == "" || typ == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		value, err := stor.GetMetric(typ, name)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write([]byte(value))
	}
}
