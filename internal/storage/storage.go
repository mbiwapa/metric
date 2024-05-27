// Package storage provides interfaces and error definitions for metric storage implementations.
// It defines common errors and can be used to abstract different storage backends.
package storage

import (
	"errors"
)

var (
	// ErrMetricNotFound is returned when a Metric is not found.
	// This error is used to indicate that a requested metric does not exist in the storage.
	ErrMetricNotFound = errors.New("metric not found")
)
