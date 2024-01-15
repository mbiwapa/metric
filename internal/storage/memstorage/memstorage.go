package memstorage

import (
	"context"
	"strconv"

	"github.com/mbiwapa/metric/internal/lib/api/format"
	"github.com/mbiwapa/metric/internal/storage"
)

// Storage Структура для хранения метрик
type Storage struct {
	Gauge   []Gauge
	Counter []Counter
}

// Gauge Структура для хранения определенного типа метрик
type Gauge struct {
	Name  string
	Value float64
}

// Counter Структура для хранения определенного типа метрик
type Counter struct {
	Name  string
	Value int64
}

// New return a new Storage instance.
func New() (*Storage, error) {
	var storage Storage
	return &storage, nil
}

// UpdateGauge saves the given Gauge metric to the memory.
func (s *Storage) UpdateGauge(ctx context.Context, key string, value float64) error {

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
func (s *Storage) UpdateCounter(ctx context.Context, key string, value int64) error {
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

// GetAllMetrics Возвращает слайс метрик 2 типов gauge и counter
func (s *Storage) GetAllMetrics(ctx context.Context) ([][]string, [][]string, error) {
	gauge := make([][]string, 0, 30)
	for _, metric := range s.Gauge {
		value := []string{metric.Name, strconv.FormatFloat(metric.Value, 'f', -1, 64)}
		gauge = append(gauge, value)
	}
	counter := make([][]string, 0, 5)
	for _, metric := range s.Counter {
		value := []string{metric.Name, strconv.FormatInt(metric.Value, 10)}
		counter = append(counter, value)
	}

	return gauge, counter, nil
}

// GetMetric Возвращает метрику по ключу
func (s *Storage) GetMetric(ctx context.Context, typ string, key string) (string, error) {
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
