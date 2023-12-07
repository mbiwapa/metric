package update

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

// Updater interface for storage
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=Updater
type Updater interface {
	GaugeUpdate(key string, value float64) error
	CounterUpdate(key string, value int64) error
}

// New returned func for update
func New(stor Updater) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		name := chi.URLParam(r, "name")
		value := chi.URLParam(r, "value")
		typ := chi.URLParam(r, "type")

		if name == "" || value == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch typ {
		case "gauge":
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err = stor.GaugeUpdate(name, val)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case "counter":
			val, err := strconv.ParseInt(value, 0, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err = stor.CounterUpdate(name, val)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			return

		}

		w.WriteHeader(http.StatusOK)
	}
}
