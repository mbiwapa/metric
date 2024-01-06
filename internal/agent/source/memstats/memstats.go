package memstats

import (
	"reflect"
	"runtime"
)

// MetricsRepo структура обертка пакета рантайм для имплементации интерфейсов агента
type MetricsRepo struct {
}

// New возвращает инстанс репы
func New() (*MetricsRepo, error) {
	var storage MetricsRepo
	return &storage, nil
}

// MetricGet возвращает значение метрики по ключу
func (s *MetricsRepo) MetricGet(metricName string, sourceType string) (float64, error) {

	ms := new(runtime.MemStats)

	runtime.ReadMemStats(ms)

	metrics := reflect.ValueOf(ms)

	val := reflect.Indirect(metrics).FieldByName(metricName)

	switch sourceType {
	case "float":
		return float64(val.Float()), nil
	case "uint":
		return float64(val.Uint()), nil
	default:
		return float64(val.Uint()), nil
	}
}
