package update

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

// CounterUpdater interface for storage
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=CounterUpdater
type CounterUpdater interface {
	CounterUpdate(key string, value int64) error
}

// New returned func for update Counter
func New(counterUpdater CounterUpdater) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		value := chi.URLParam(r, "value")

		if name == "" || value == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		val, err := strconv.ParseInt(value, 0, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = counterUpdater.CounterUpdate(name, val)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
