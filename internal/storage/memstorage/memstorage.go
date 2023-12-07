package memstorage

import (
	"errors"
	"fmt"
	"strconv"
)

// MemStorage Структура для хранения метрик
type MemStorage struct {
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

// New return a new MemStorage instance.
func New() (*MemStorage, error) {
	var storage MemStorage
	return &storage, nil
}

// GaugeUpdate saves the given Gauge metric to the memory.
func (s *MemStorage) GaugeUpdate(key string, value float64) error {

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

// CounterUpdate saves the given Counter metric to the memory.
func (s *MemStorage) CounterUpdate(key string, value int64) error {
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
	fmt.Println(s)

	return nil
}

// GetAllMetrics Возвращает слайс метрик 2 типов gauge и counter
func (s *MemStorage) GetAllMetrics() ([][]string, [][]string, error) {
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
func (s *MemStorage) GetMetric(typ string, key string) (string, error) {
	if typ == "gauge" {
		for _, metric := range s.Gauge {
			if metric.Name == key {
				return strconv.FormatFloat(metric.Value, 'f', -1, 64), nil
			}
		}
	}
	if typ == "counter" {
		for _, metric := range s.Counter {
			if metric.Name == key {
				return strconv.FormatInt(metric.Value, 10), nil
			}
		}
	}

	return "", errors.New("metric not found")
}
