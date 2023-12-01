package update

import (
	"net/http"
	"strconv"
	"strings"
)

// CounterUpdater interface for storage
type CounterUpdater interface {
	CounterUpdate(key string, value int64) error
}

// New returned func for update Counter
func New(counterUpdater CounterUpdater) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// if r.Method != http.MethodPost {
		// 	w.WriteHeader(http.StatusMethodNotAllowed)
		// return
		// }
		path := strings.TrimPrefix(r.URL.Path, "/update/counter/")
		params := strings.Split(path, "/")

		if len(params) < 2 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		val, err := strconv.ParseInt(params[1], 0, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = counterUpdater.CounterUpdate(params[0], val)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
