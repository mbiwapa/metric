package memstatssource

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

// GetObservableMetrics возвращает список метрик для отслеживание агентом
func (s *MetricsRepo) GetObservableMetrics() (map[string]string, error) {

	observableMetrics := map[string]string{
		"Frees":         "uint",
		"Alloc":         "uint",
		"BuckHashSys":   "uint",
		"GCCPUFraction": "float",
		"GCSys":         "uint",
		"HeapAlloc":     "uint",
		"HeapIdle":      "uint",
		"HeapInuse":     "uint",
		"HeapObjects":   "uint",
		"HeapReleased":  "uint",
		"HeapSys":       "uint",
		"LastGC":        "uint",
		"Lookups":       "uint",
		"MCacheInuse":   "uint",
		"MCacheSys":     "uint",
		"MSpanInuse":    "uint",
		"MSpanSys":      "uint",
		"Mallocs":       "uint",
		"NextGC":        "uint",
		"NumForcedGC":   "uint",
		"NumGC":         "uint",
		"OtherSys":      "uint",
		"PauseTotalNs":  "uint",
		"StackInuse":    "uint",
		"StackSys":      "uint",
		"Sys":           "uint",
		"TotalAlloc":    "uint",
	}
	return observableMetrics, nil
}
