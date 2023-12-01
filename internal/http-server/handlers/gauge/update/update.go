package update

import (
	"net/http"
	"strconv"
	"strings"
)

// GaugeUpdater interface for storage
type GaugeUpdater interface {
	GaugeUpdate(key string, value float64) error
}

// New returned func for update Gauge
func New(gaugeUpdater GaugeUpdater) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/update/gauge/")
		params := strings.Split(path, "/")

		if len(params) < 2 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		val, err := strconv.ParseFloat(params[1], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = gaugeUpdater.GaugeUpdate(params[0], val)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
