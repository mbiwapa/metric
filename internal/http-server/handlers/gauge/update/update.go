package update

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

// GaugeUpdater interface for storage
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=GaugeUpdater
type GaugeUpdater interface {
	GaugeUpdate(key string, value float64) error
}

// New returned func for update Gauge
func New(gaugeUpdater GaugeUpdater) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		name := chi.URLParam(r, "name")
		value := chi.URLParam(r, "value")

		if name == "" || value == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = gaugeUpdater.GaugeUpdate(name, val)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
