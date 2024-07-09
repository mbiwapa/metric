// Package storage provides interfaces and error definitions for metric storage implementations.
// It defines common errors and can be used to abstract different storage backends.
package storage

import (
	"context"
	"errors"
)

var (
	// ErrMetricNotFound is returned when a Metric is not found.
	// This error is used to indicate that a requested metric does not exist in the storage.
	ErrMetricNotFound = errors.New("metric not found")
)

// Storage is an interface that defines the methods required for a storage implementation.
type Storage interface {
	UpdateGauge(ctx context.Context, key string, value float64) error
	UpdateCounter(ctx context.Context, key string, value int64) error
	GetAllMetrics(ctx context.Context) ([][]string, [][]string, error)
	GetMetric(ctx context.Context, typ string, key string) (string, error)
	UpdateBatch(ctx context.Context, gauges [][]string, counters [][]string) error
	Ping(ctx context.Context) error
	Close()
}
