package update

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/mbiwapa/metric/internal/lib/api/format"
)

// Updater interface for storage
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=Updater
type Updater interface {
	UpdateGauge(key string, value float64) error
	UpdateCounter(key string, value int64) error
}

// Backuper interface for backuper
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=Backuper
type Backuper interface {
	SaveToStruct(typ string, name string, value string) error
	SaveToFile()
	IsSyncMode() bool
}

// New returned func for update
func New(log *zap.Logger, storage Updater, backup Backuper) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.update.New"

		log.With(
			zap.String("op", op),
			zap.String("request_id", middleware.GetReqID(r.Context())),
		)

		name := chi.URLParam(r, "name")
		typ := chi.URLParam(r, "type")

		// value := chi.URLParam(r, "value")
		//TODO от chi не работает, режет по . и не принимает float
		value := strings.TrimPrefix(r.URL.Path, "/update/"+typ+"/"+name+"/")

		if name == "" || value == "" {
			log.Error(
				"Name or Value is empty!",
				zap.String("name", name),
				zap.String("value", value))

			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch typ {
		case format.Gauge:
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				log.Error("Failed to parse gauge value", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err = storage.UpdateGauge(name, val)
			if err != nil {
				log.Error("Failed to update gauge value", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case format.Counter:
			val, err := strconv.ParseInt(value, 0, 64)
			if err != nil {
				log.Error("Failed to parse counter value", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err = storage.UpdateCounter(name, val)
			if err != nil {
				log.Error("Failed to update counter value", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		default:
			log.Error("Undefined metric type", zap.String("type", typ))
			w.WriteHeader(http.StatusBadRequest)
			return

		}

		if backup.IsSyncMode() {
			backup.SaveToStruct(typ, name, value)
			backup.SaveToFile()
		}

		w.WriteHeader(http.StatusOK)
	}
}
