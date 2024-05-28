// Package memstatssource provides a source that retrieves metrics using the runtime package's MemStats.
// It provides functions to retrieve the value of a metric by its name and type and to retrieve a list of observable metrics.
package memstatssource

import (
	"reflect"
	"runtime"
)

// MetricsRepo is a wrapper structure for the runtime package to implement agent interfaces.
type MetricsRepo struct {
}

// New returns an instance of MetricsRepo.
// It initializes and returns a new MetricsRepo object.
func New() (*MetricsRepo, error) {
	var storage MetricsRepo
	return &storage, nil
}

// MetricGet returns the value of a metric by its name and type.
// Parameters:
// - metricName: the name of the metric to retrieve.
// - sourceType: the type of the metric value ("float" or "uint").
// Returns:
// - float64: the value of the requested metric.
// - error: an error if the metric retrieval fails.
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

// GetObservableMetrics returns a list of metrics to be monitored by the agent.
// Returns:
// - map[string]string: a map where the key is the metric name and the value is the type of the metric ("float" or "uint").
// - error: an error if the retrieval of observable metrics fails.
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
