// Package memstorage provides an in-memory storage implementation for metrics.
// It includes methods for creating, updating, and retrieving Gauge and Counter metrics.
package memstorage

import (
	"context"
	"strconv"

	"github.com/mbiwapa/metric/internal/lib/api/format"
	"github.com/mbiwapa/metric/internal/storage"
)

// Storage is a structure for storing metrics.
// It contains slices of Gauge and Counter metrics.
type Storage struct {
	Gauge   []Gauge   // Slice of Gauge metrics
	Counter []Counter // Slice of Counter metrics
}

// Gauge is a structure for storing a specific type of metric.
// It contains the name and value of the gauge metric.
type Gauge struct {
	Name  string  // Name of the gauge metric
	Value float64 // Value of the gauge metric
}

// Counter is a structure for storing a specific type of metric.
// It contains the name and value of the counter metric.
type Counter struct {
	Name  string // Name of the counter metric
	Value int64  // Value of the counter metric
}

// New creates and returns a new instance of Storage.
// This function initializes a new Storage struct, which is used to store metrics in memory.
// Returns:
// - *Storage: a pointer to the newly created Storage instance.
// - error: always returns nil as there are no error conditions in this function.
func New() (*Storage, error) {
	var storage Storage
	return &storage, nil
}

// UpdateGauge saves the given Gauge metric to the memory.
// Parameters:
// - ctx: context for managing request-scoped values, cancelation, and deadlines.
// - key: the name of the gauge metric.
// - value: the value of the gauge metric.
// Returns:
// - error: if any error occurs during the update.
func (s *Storage) UpdateGauge(_ context.Context, key string, value float64) error {

	changed := false
	for i := 0; i < len(s.Gauge); i++ {
		if s.Gauge[i].Name == key {
			s.Gauge[i].Value = value
			changed = true
		}
	}

	if !changed {
		var metric Gauge
		metric.Name = key
		metric.Value = value
		s.Gauge = append(s.Gauge, metric)
	}

	return nil
}

// UpdateCounter saves the given Counter metric to the memory.
// Parameters:
// - ctx: context for managing request-scoped values, cancelation, and deadlines.
// - key: the name of the counter metric.
// - value: the value of the counter metric.
// Returns:
// - error: if any error occurs during the update.
func (s *Storage) UpdateCounter(_ context.Context, key string, value int64) error {
	changed := false
	for i := 0; i < len(s.Counter); i++ {
		if s.Counter[i].Name == key {
			s.Counter[i].Value = int64(value) + s.Counter[i].Value
			changed = true
		}
	}

	if !changed {
		var metric Counter
		metric.Name = key
		metric.Value = int64(value)
		s.Counter = append(s.Counter, metric)
	}
	return nil
}

// GetAllMetrics returns slices of metrics of two types: gauge and counter.
// Parameters:
// - ctx: context for managing request-scoped values, cancelation, and deadlines.
// Returns:
// - [][]string: slice of gauge metrics, where each metric is represented as a slice of strings [name, value].
// - [][]string: slice of counter metrics, where each metric is represented as a slice of strings [name, value].
// - error: if any error occurs during the retrieval.
func (s *Storage) GetAllMetrics(_ context.Context) ([][]string, [][]string, error) {
	gauge := make([][]string, 0, 40)
	for _, metric := range s.Gauge {
		value := []string{metric.Name, strconv.FormatFloat(metric.Value, 'f', -1, 64)}
		gauge = append(gauge, value)
	}
	counter := make([][]string, 0, 2)
	for _, metric := range s.Counter {
		value := []string{metric.Name, strconv.FormatInt(metric.Value, 10)}
		counter = append(counter, value)
	}

	return gauge, counter, nil
}

// GetMetric returns a metric by key.
// Parameters:
// - ctx: context for managing request-scoped values, cancelation, and deadlines.
// - typ: the type of the metric (gauge or counter).
// - key: the name of the metric.
// Returns:
// - string: the value of the metric as a string.
// - error: if the metric is not found or any other error occurs.
func (s *Storage) GetMetric(_ context.Context, typ string, key string) (string, error) {
	if typ == format.Gauge {
		for _, metric := range s.Gauge {
			if metric.Name == key {
				return strconv.FormatFloat(metric.Value, 'f', -1, 64), nil
			}
		}
	}
	if typ == format.Counter {
		for _, metric := range s.Counter {
			if metric.Name == key {
				return strconv.FormatInt(metric.Value, 10), nil
			}
		}
	}

	return "", storage.ErrMetricNotFound
}

// UpdateBatch saves the given Gauge and Counter metrics to the memory.
// Parameters:
// - ctx: context for managing request-scoped values, cancelation, and deadlines.
// - gauges: slice of gauge metrics, where each metric is represented as a slice of strings [name, value].
// - counters: slice of counter metrics, where each metric is represented as a slice of strings [name, value].
// Returns:
// - error: if any error occurs during the update.
func (s *Storage) UpdateBatch(_ context.Context, gauges [][]string, counters [][]string) error {
	for _, gauge := range gauges {
		changed := false
		for i := 0; i < len(s.Gauge); i++ {
			if s.Gauge[i].Name == gauge[0] {
				val, err := strconv.ParseFloat(gauge[1], 64)
				if err != nil {
					return err
				}
				s.Gauge[i].Value = val
				changed = true
			}
		}

		if !changed {
			var metric Gauge
			metric.Name = gauge[0]
			val, err := strconv.ParseFloat(gauge[1], 64)
			if err != nil {
				return err
			}
			metric.Value = val
			s.Gauge = append(s.Gauge, metric)
		}

	}

	for _, counter := range counters {

		changed := false
		for i := 0; i < len(s.Counter); i++ {
			if s.Counter[i].Name == counter[0] {
				val, err := strconv.ParseInt(counter[1], 0, 64)
				if err != nil {
					return err
				}
				s.Counter[i].Value = val + s.Counter[i].Value
				changed = true
			}
		}

		if !changed {
			var metric Counter
			metric.Name = counter[0]
			val, err := strconv.ParseInt(counter[1], 0, 64)
			if err != nil {
				return err
			}
			metric.Value = int64(val)
			s.Counter = append(s.Counter, metric)
		}
	}

	return nil
}
